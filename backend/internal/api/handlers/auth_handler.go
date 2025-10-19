package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/audit"
	"github.com/pixsaas/backend/internal/repository"
	"github.com/pixsaas/backend/internal/security"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler gerencia autenticação
type AuthHandler struct {
	db           *gorm.DB
	jwtService   *security.JWTService
	auditService *audit.AuditService
	userRepo     *repository.UserRepository
}

// NewAuthHandler cria um novo handler de autenticação
func NewAuthHandler(db *gorm.DB, jwtService *security.JWTService, auditService *audit.AuditService) *AuthHandler {
	return &AuthHandler{
		db:           db,
		jwtService:   jwtService,
		auditService: auditService,
		userRepo:     repository.NewUserRepository(db),
	}
}

// LoginRequest representa uma requisição de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginResponse representa uma resposta de login
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo representa informações do usuário
type UserInfo struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	MerchantID *uuid.UUID `json:"merchant_id,omitempty"`
}

// Login autentica um usuário
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	
	// Buscar usuário
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil {
		_ = h.auditService.LogAuthentication(c.Context(), req.Email, c.IP(), false, "user not found")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}
	
	// Verificar se usuário está ativo
	if !user.Active {
		_ = h.auditService.LogAuthentication(c.Context(), req.Email, c.IP(), false, "user inactive")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user is inactive",
		})
	}
	
	// Verificar senha
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		_ = h.auditService.LogAuthentication(c.Context(), req.Email, c.IP(), false, "invalid password")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}
	
	// Gerar tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(user.ID, user.MerchantID, user.Email, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}
	
	// Atualizar last_login
	now := time.Now()
	user.LastLogin = &now
	if err := h.userRepo.Update(c.Context(), user); err != nil {
		// Log error but don't fail the login
		log.Printf("Warning: Failed to update last_login: %v", err)
	}
	
	// Salvar refresh token no banco
	// TODO: Implementar salvamento de refresh token
	
	// Log de sucesso
	_ = h.auditService.LogAuthentication(c.Context(), req.Email, c.IP(), true, "")
	
	return c.JSON(LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		ExpiresAt:    tokenPair.ExpiresAt,
		User: UserInfo{
			ID:         user.ID,
			Email:      user.Email,
			Name:       user.Name,
			Role:       string(user.Role),
			MerchantID: user.MerchantID,
		},
	})
}

// RefreshTokenRequest representa uma requisição de refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken renova o access token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	
	// Validar refresh token
	userID, err := h.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid refresh token",
		})
	}
	
	// Buscar usuário
	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not found",
		})
	}
	
	// Verificar se usuário está ativo
	if !user.Active {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user is inactive",
		})
	}
	
	// Gerar novos tokens
	tokenPair, err := h.jwtService.GenerateTokenPair(user.ID, user.MerchantID, user.Email, string(user.Role))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}
	
	return c.JSON(LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		ExpiresAt:    tokenPair.ExpiresAt,
		User: UserInfo{
			ID:         user.ID,
			Email:      user.Email,
			Name:       user.Name,
			Role:       string(user.Role),
			MerchantID: user.MerchantID,
		},
	})
}

// Logout invalida o refresh token
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// TODO: Implementar revogação de refresh token
	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

// Me retorna informações do usuário autenticado
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "user not authenticated",
		})
	}
	
	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}
	
	return c.JSON(UserInfo{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		Role:       string(user.Role),
		MerchantID: user.MerchantID,
	})
}
