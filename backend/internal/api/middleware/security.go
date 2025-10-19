package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SecurityHeaders adiciona headers de segurança
func SecurityHeaders() fiber.Handler {
	return helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		HSTSMaxAge:                31536000,
		HSTSExcludeSubdomains:     false,
		ContentSecurityPolicy:     "default-src 'self'",
		ReferrerPolicy:            "no-referrer",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=()",
	})
}

// CORS configura CORS
func CORS(allowedOrigins []string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     joinOrigins(allowedOrigins),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key",
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

// Recover recupera de panics
func Recover() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
	})
}

func joinOrigins(origins []string) string {
	if len(origins) == 0 {
		return "*"
	}
	
	result := origins[0]
	for i := 1; i < len(origins); i++ {
		result += "," + origins[i]
	}
	return result
}

// IPWhitelist verifica se o IP está na whitelist do merchant
func IPWhitelist() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Implementar verificação de IP whitelist
		// Buscar merchant do contexto e verificar se IP está na whitelist
		return c.Next()
	}
}
