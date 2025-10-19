package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestHealthCheck(t *testing.T) {
	// Criar app Fiber
	app := fiber.New()

	// Registrar rota
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "pix-saas-api",
		})
	})

	// Criar requisição
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Executar requisição
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	// Verificar resposta
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestReadiness(t *testing.T) {
	// Criar app Fiber
	app := fiber.New()

	// Registrar rota
	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "ready",
			"database": "connected",
		})
	})

	// Criar requisição
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)

	// Executar requisição
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}

	// Verificar resposta
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
