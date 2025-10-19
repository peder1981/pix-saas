package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/security"
)

// AuthMiddleware valida JWT tokens
func AuthMiddleware(jwtService *security.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}
		
		token, err := security.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization header",
			})
		}
		
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}
		
		// Armazenar claims no contexto
		c.Locals("user_id", claims.UserID)
		c.Locals("merchant_id", claims.MerchantID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		
		return c.Next()
	}
}

// APIKeyMiddleware valida API Keys
func APIKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing API key",
			})
		}
		
		// TODO: Validar API key no banco de dados
		// Por enquanto, apenas verificar se existe
		if !strings.HasPrefix(apiKey, "pk_") && !strings.HasPrefix(apiKey, "sk_") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid API key format",
			})
		}
		
		// Armazenar merchant_id no contexto após validação
		// c.Locals("merchant_id", merchantID)
		
		return c.Next()
	}
}

// RequireRole verifica se o usuário tem a role necessária
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "role not found in context",
			})
		}
		
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}
		
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient permissions",
		})
	}
}

// RequireMerchant verifica se o usuário pertence ao merchant
func RequireMerchant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		merchantID, ok := c.Locals("merchant_id").(*uuid.UUID)
		if !ok || merchantID == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "merchant context required",
			})
		}
		
		return c.Next()
	}
}
