package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/pixsaas/backend/configs"
	"github.com/pixsaas/backend/internal/api/handlers"
	"github.com/pixsaas/backend/internal/api/middleware"
	"github.com/pixsaas/backend/internal/audit"
	"github.com/pixsaas/backend/internal/domain"
	"github.com/pixsaas/backend/internal/providers"
	"github.com/pixsaas/backend/internal/providers/bb"
	"github.com/pixsaas/backend/internal/providers/inter"
	"github.com/pixsaas/backend/internal/providers/santander"
	"github.com/pixsaas/backend/internal/security"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	// Carregar configuração
	cfg, err := configs.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Conectar ao banco de dados
	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Auto-migrate (apenas em desenvolvimento)
	if cfg.Server.IsDevelopment() {
		if migrateErr := db.AutoMigrate(
			&domain.Merchant{},
			&domain.User{},
			&domain.Provider{},
			&domain.MerchantProvider{},
			&domain.Transaction{},
			&domain.AuditLog{},
			&domain.Webhook{},
			&domain.WebhookDelivery{},
			&domain.APIKey{},
			&domain.RefreshToken{},
		); migrateErr != nil {
			log.Printf("Aviso: Erro no auto-migrate: %v", migrateErr)
		}
	}

	// Inicializar serviços
	encryptionKey, err := base64.StdEncoding.DecodeString(cfg.Encryption.Key)
	if err != nil || len(encryptionKey) != 32 {
		log.Fatalf("Chave de criptografia inválida. Deve ser 32 bytes em base64")
	}

	encryptionService, err := security.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatalf("Erro ao criar serviço de criptografia: %v", err)
	}

	jwtService := security.NewJWTService(
		[]byte(cfg.JWT.SecretKey),
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	auditService := audit.NewAuditService(db)

	// Registrar providers
	providerRegistry := providers.NewProviderRegistry()
	// TODO: Atualizar Bradesco e Itaú para nova interface
	// providerRegistry.Register(bradesco.NewBradescoProvider())
	// providerRegistry.Register(itau.NewItauProvider())
	providerRegistry.Register(bb.NewBBProvider())
	providerRegistry.Register(santander.NewSantanderProvider())
	providerRegistry.Register(inter.NewInterProvider())

	// Criar aplicação Fiber
	app := fiber.New(fiber.Config{
		AppName:      "PIX SaaS API",
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		ErrorHandler: customErrorHandler,
	})

	// Middlewares globais
	app.Use(logger.New())
	app.Use(middleware.Recover())
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.CORS(cfg.Server.AllowedOrigins))

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.Server.RateLimitRPS)
	app.Use(rateLimiter.Middleware())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1
	v1 := app.Group("/v1")

	// Rotas públicas
	authHandler := handlers.NewAuthHandler(db, jwtService, auditService)
	v1.Post("/auth/login", authHandler.Login)
	v1.Post("/auth/refresh", authHandler.RefreshToken)

	// Rotas autenticadas (JWT)
	authenticated := v1.Group("")
	authenticated.Use(middleware.AuthMiddleware(jwtService))
	authenticated.Use(middleware.AuditMiddleware(auditService))

	authenticated.Get("/auth/me", authHandler.Me)
	authenticated.Post("/auth/logout", authHandler.Logout)

	// Rotas de transações (requer merchant)
	txHandler := handlers.NewTransactionHandler(db, auditService, encryptionService, providerRegistry)
	transactions := authenticated.Group("/transactions")
	transactions.Use(middleware.RequireMerchant())

	transactions.Post("/transfer", txHandler.CreateTransfer)
	transactions.Get("/:id", txHandler.GetTransaction)
	transactions.Get("", txHandler.ListTransactions)

	// Rotas administrativas
	admin := authenticated.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))
	// TODO: Adicionar rotas administrativas

	// Iniciar servidor
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🚀 Servidor iniciando em http://localhost%s", addr)
	log.Printf("📝 Ambiente: %s", cfg.Server.Environment)
	log.Printf("🏦 Providers registrados: %d", len(providerRegistry.GetAll()))

	// Graceful shutdown
	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// Aguardar sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Desligando servidor...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Erro ao desligar servidor: %v", err)
		shutdownCancel()
		os.Exit(1)
	}

	log.Println("✅ Servidor desligado com sucesso")
}

func connectDatabase(cfg *configs.Config) (*gorm.DB, error) {
	logLevel := gormlogger.Silent
	if cfg.Server.IsDevelopment() {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}
