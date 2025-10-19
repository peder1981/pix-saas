# Guia de Instalação - PIX SaaS

## 📋 Pré-requisitos

### Desenvolvimento Local
- Go 1.21 ou superior
- PostgreSQL 15+
- Make (opcional, mas recomendado)
- Docker e Docker Compose (opcional)

### Produção
- Docker e Docker Compose
- Servidor Linux (Ubuntu 22.04 LTS recomendado)
- Mínimo 2GB RAM, 2 vCPUs
- 20GB de disco

## 🚀 Instalação Rápida (Docker)

### 1. Clone o repositório

```bash
git clone https://github.com/peder1981/pix-saas.git
cd pix-saas
```

### 2. Configure variáveis de ambiente

```bash
# Copiar exemplo
cp backend/.env.example .env

# Gerar chaves de segurança
openssl rand -base64 32  # ENCRYPTION_KEY
openssl rand -base64 64  # JWT_SECRET_KEY

# Editar .env com as chaves geradas
nano .env
```

### 3. Inicie os containers

```bash
docker-compose up -d
```

### 4. Verifique a saúde da API

```bash
curl http://localhost:8080/health
```

Resposta esperada:
```json
{
  "status": "healthy",
  "time": "2024-01-20T10:00:00Z"
}
```

## 💻 Instalação para Desenvolvimento

### 1. Instale as dependências Go

```bash
cd backend
go mod download
```

### 2. Configure o banco de dados

```bash
# Criar banco
createdb pixsaas

# Executar migrations
psql -d pixsaas -f migrations/001_initial_schema.sql
```

Ou usando Make:
```bash
make migrate-up DB_NAME=pixsaas
```

### 3. Configure variáveis de ambiente

```bash
cp .env.example .env
```

Edite o arquivo `.env`:
```env
# Server
PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=pixsaas
DB_SSLMODE=disable

# JWT (gere com: openssl rand -base64 64)
JWT_SECRET_KEY=sua-chave-jwt-aqui

# Encryption (gere com: openssl rand -base64 32)
ENCRYPTION_KEY=sua-chave-criptografia-aqui

# Rate Limiting
RATE_LIMIT_RPS=100

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### 4. Execute a aplicação

```bash
# Usando Make
make run

# Ou diretamente
go run cmd/api/main.go
```

A API estará disponível em `http://localhost:8080`

## 🔧 Configuração Inicial

### 1. Adicionar Providers via CLI

```bash
# Compilar CLI
go build -o pixsaas-cli cmd/cli/main.go

# Adicionar Bradesco
./pixsaas-cli provider add \
  --code bradesco \
  --name "Bradesco" \
  --ispb "60746948" \
  --type bank \
  --base-url "https://qrpix.bradesco.com.br" \
  --auth-url "https://qrpix.bradesco.com.br/oauth/token"

# Adicionar Itaú
./pixsaas-cli provider add \
  --code itau \
  --name "Itaú Unibanco" \
  --ispb "60701190" \
  --type bank \
  --base-url "https://api.itau.com.br" \
  --auth-url "https://sts.itau.com.br/api/oauth/token"

# Listar providers
./pixsaas-cli provider list
```

### 2. Criar Usuário Admin (via SQL)

```sql
-- Gerar hash de senha (use bcrypt)
-- Senha: Admin123!
INSERT INTO users (id, email, password, name, role, active)
VALUES (
  gen_random_uuid(),
  'admin@pixsaas.com.br',
  '$2a$10$exemplo_hash_bcrypt_aqui',
  'Administrador',
  'admin',
  true
);
```

### 3. Criar Merchant de Teste

```sql
INSERT INTO merchants (id, name, document, email, api_key, active)
VALUES (
  gen_random_uuid(),
  'Merchant Teste',
  '12345678000190',
  'merchant@teste.com.br',
  'pk_test_' || encode(gen_random_bytes(32), 'hex'),
  true
);
```

## 🧪 Testando a API

### 1. Login

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@pixsaas.com.br",
    "password": "Admin123!"
  }'
```

Resposta:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "token_type": "Bearer",
  "expires_in": 900,
  "expires_at": "2024-01-20T10:15:00Z",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "email": "admin@pixsaas.com.br",
    "name": "Administrador",
    "role": "admin"
  }
}
```

### 2. Criar Transferência PIX

```bash
TOKEN="seu-access-token-aqui"

curl -X POST http://localhost:8080/v1/transactions/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "external_id": "ORDER-12345",
    "amount": 10000,
    "description": "Pagamento teste",
    "provider_code": "bradesco",
    "payee_name": "João Silva",
    "payee_document": "12345678901",
    "payee_pix_key": "joao@example.com",
    "payee_pix_key_type": "email"
  }'
```

### 3. Consultar Transação

```bash
curl -X GET http://localhost:8080/v1/transactions/{id} \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Listar Transações

```bash
curl -X GET "http://localhost:8080/v1/transactions?limit=10&offset=0&status=completed" \
  -H "Authorization: Bearer $TOKEN"
```

## 🐳 Docker em Produção

### 1. Configurar variáveis de ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
JWT_SECRET_KEY=sua-chave-segura-producao
ENCRYPTION_KEY=sua-chave-criptografia-producao
```

### 2. Build e deploy

```bash
# Build
docker-compose build

# Deploy
docker-compose up -d

# Verificar logs
docker-compose logs -f api
```

### 3. Configurar proxy reverso (Nginx)

```nginx
server {
    listen 80;
    server_name api.pixsaas.com.br;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 4. Configurar SSL com Let's Encrypt

```bash
sudo certbot --nginx -d api.pixsaas.com.br
```

## 📊 Monitoramento (Opcional)

### Habilitar Prometheus e Grafana

```bash
docker-compose --profile monitoring up -d
```

Acessar:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin)

## 🔒 Segurança em Produção

### Checklist de Segurança

- [ ] Alterar todas as senhas padrão
- [ ] Gerar novas chaves JWT e Encryption
- [ ] Configurar SSL/TLS (HTTPS)
- [ ] Configurar firewall (permitir apenas portas 80, 443)
- [ ] Habilitar backups automáticos do PostgreSQL
- [ ] Configurar rate limiting adequado
- [ ] Revisar logs de auditoria regularmente
- [ ] Implementar rotação de logs
- [ ] Configurar alertas de segurança
- [ ] Testar disaster recovery

## 🛠️ Comandos Úteis

### Make

```bash
# Executar aplicação
make run

# Compilar
make build

# Testes
make test

# Cobertura de testes
make test-coverage

# Limpar builds
make clean

# Formatar código
make fmt

# Linter
make lint

# Gerar chaves
make generate-key
make generate-jwt-secret
```

### Docker

```bash
# Iniciar
docker-compose up -d

# Parar
docker-compose down

# Ver logs
docker-compose logs -f

# Rebuild
docker-compose build --no-cache

# Limpar volumes
docker-compose down -v
```

### CLI

```bash
# Listar providers
./pixsaas-cli provider list

# Listar merchants
./pixsaas-cli merchant list

# Gerar chave de criptografia
./pixsaas-cli keys generate
```

## 🐛 Troubleshooting

### Erro de conexão com banco de dados

```bash
# Verificar se PostgreSQL está rodando
sudo systemctl status postgresql

# Verificar conexão
psql -h localhost -U postgres -d pixsaas
```

### Erro de permissão

```bash
# Dar permissão ao usuário
sudo -u postgres psql
GRANT ALL PRIVILEGES ON DATABASE pixsaas TO postgres;
```

### Porta já em uso

```bash
# Verificar processo usando a porta
lsof -i :8080

# Matar processo
kill -9 <PID>
```

### Logs da aplicação

```bash
# Docker
docker-compose logs -f api

# Local
tail -f logs/app.log
```

## 📚 Próximos Passos

Após a instalação:

1. Leia a [documentação da API](./docs/api/openapi.yaml)
2. Configure webhooks para receber notificações
3. Adicione suas credenciais bancárias
4. Teste em ambiente sandbox
5. Configure monitoramento
6. Implemente backup automático

## 🆘 Suporte

- Email: suporte@pixsaas.com.br
- Documentação: https://docs.pixsaas.com.br
- Issues: https://github.com/peder1981/pix-saas/issues
