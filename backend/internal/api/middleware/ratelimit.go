package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RateLimiter implementa rate limiting simples em memória
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    requestsPerSecond,
		window:   time.Second,
	}

	// Limpar entradas antigas periodicamente
	go rl.cleanup()

	return rl
}

// Middleware retorna o middleware de rate limiting
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Usar IP + API Key como identificador
		identifier := c.IP()
		if apiKey := c.Get("X-API-Key"); apiKey != "" {
			identifier = apiKey
		}

		if !rl.allow(identifier) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": 1,
			})
		}

		return c.Next()
	}
}

func (rl *RateLimiter) allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Obter requisições do identificador
	requests := rl.requests[identifier]

	// Filtrar requisições dentro da janela
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Verificar se excedeu o limite
	if len(validRequests) >= rl.limit {
		rl.requests[identifier] = validRequests
		return false
	}

	// Adicionar nova requisição
	validRequests = append(validRequests, now)
	rl.requests[identifier] = validRequests

	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window * 2)

		for identifier, requests := range rl.requests {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, identifier)
			} else {
				rl.requests[identifier] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}
