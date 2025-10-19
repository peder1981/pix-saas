package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pixsaas/backend/internal/audit"
)

// AuditMiddleware registra todas as requisições
func AuditMiddleware(auditService *audit.AuditService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Processar requisição
		err := c.Next()
		
		// Calcular duração
		duration := time.Since(start).Milliseconds()
		
		// Extrair informações do contexto
		var merchantID *uuid.UUID
		if mid, ok := c.Locals("merchant_id").(*uuid.UUID); ok {
			merchantID = mid
		}
		
		// Registrar log de auditoria de forma assíncrona
		go func() {
			if merchantID != nil {
				auditService.LogAPIAccess(
					c.Context(),
					*merchantID,
					c.Method(),
					c.Path(),
					c.IP(),
					c.Get("User-Agent"),
					c.Response().StatusCode(),
					duration,
				)
			}
		}()
		
		return err
	}
}
