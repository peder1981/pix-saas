.PHONY: help run build test clean migrate-up migrate-down docker-up docker-down

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Executa a aplicação
	go run cmd/api/main.go

build: ## Compila a aplicação
	go build -o bin/api cmd/api/main.go

test: ## Executa os testes
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## Executa testes e gera relatório de cobertura
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Relatório de cobertura gerado em coverage.html"

clean: ## Remove arquivos compilados
	rm -rf bin/
	rm -f coverage.txt coverage.html

deps: ## Instala dependências
	go mod download
	go mod tidy

migrate-up: ## Executa migrations
	psql -d $(DB_NAME) -f migrations/001_initial_schema.sql

migrate-down: ## Reverte migrations
	@echo "⚠️  Atenção: Isso irá remover todas as tabelas!"
	@read -p "Tem certeza? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		psql -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
	fi

docker-up: ## Inicia containers Docker
	docker-compose up -d

docker-down: ## Para containers Docker
	docker-compose down

docker-logs: ## Mostra logs dos containers
	docker-compose logs -f

lint: ## Executa linter
	golangci-lint run

fmt: ## Formata código
	go fmt ./...
	goimports -w .

generate-key: ## Gera chave de criptografia
	@echo "Chave de criptografia (adicione ao .env como ENCRYPTION_KEY):"
	@openssl rand -base64 32

generate-jwt-secret: ## Gera secret JWT
	@echo "Secret JWT (adicione ao .env como JWT_SECRET_KEY):"
	@openssl rand -base64 64

dev: ## Executa em modo desenvolvimento com hot reload
	air

.DEFAULT_GOAL := help
