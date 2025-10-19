# PIX SaaS Platform

## 🚀 Visão Geral

Plataforma SaaS robusta, escalável e segura para pagamentos via PIX, totalmente plug and play, compatível com múltiplas instituições financeiras brasileiras.

## 🎯 Características Principais

- ✅ **Multi-tenant**: Isolamento completo de dados entre merchants
- ✅ **Plug and Play**: Integração simplificada com APIs normalizadas
- ✅ **Multi-banco**: Suporte para 18+ instituições financeiras brasileiras
- ✅ **Segurança PCI DSS**: Criptografia end-to-end, TLS 1.3, AES-256
- ✅ **Auditoria Completa**: Logs de todas as operações com retenção de 5 anos
- ✅ **Alta Disponibilidade**: Arquitetura escalável e resiliente
- ✅ **Dashboard**: Interface completa para merchants
- ✅ **CLI Admin**: Ferramenta para gestão de providers

## 🏗️ Arquitetura

### Backend (Go)
- **Framework**: Fiber (alta performance)
- **Banco de Dados**: PostgreSQL 15+ com replicação
- **Autenticação**: JWT com refresh tokens
- **Padrão**: Clean Architecture
- **Criptografia**: AES-256-GCM para dados sensíveis

### Frontend (Next.js)
- **Framework**: Next.js 14+ (App Router)
- **UI**: TailwindCSS + shadcn/ui
- **Gráficos**: Recharts
- **Estado**: React Query

### CLI (Go)
- **Framework**: Cobra CLI
- **Função**: Gestão de providers e configurações

## 🏦 Instituições Suportadas

### Bancos Tradicionais
- Bradesco
- Itaú
- Banco do Brasil
- Santander
- Caixa Econômica Federal
- Safra
- BTG Pactual
- Banco Original

### Bancos Digitais
- Nubank
- Inter
- C6 Bank
- Next
- Neon

### Cooperativas
- Sicoob
- Sicredi
- Unicred

### Fintechs e PSPs
- PagSeguro/PagBank
- Mercado Pago
- PicPay
- Stone
- Cielo

## 📁 Estrutura do Projeto

```
pix-saas/
├── backend/                 # Backend Go
│   ├── cmd/                # Entry points
│   │   ├── api/           # API server
│   │   └── cli/           # CLI admin tool
│   ├── configs/           # Configurações
│   ├── internal/          # Código interno
│   │   ├── api/          # HTTP handlers
│   │   ├── domain/       # Modelos de domínio
│   │   ├── usecases/     # Casos de uso
│   │   ├── providers/    # Integrações bancárias
│   │   ├── repository/   # Camada de dados
│   │   ├── security/     # Criptografia e segurança
│   │   ├── audit/        # Sistema de auditoria
│   │   └── infrastructure/ # Infraestrutura
│   ├── migrations/        # Migrações DB
│   └── pkg/              # Pacotes públicos
├── frontend/              # Dashboard Next.js
│   ├── app/              # App Router
│   ├── components/       # Componentes React
│   ├── lib/             # Utilitários
│   └── public/          # Assets estáticos
├── docs/                 # Documentação
│   ├── api/             # OpenAPI/Swagger
│   ├── architecture/    # Diagramas
│   └── guides/          # Guias de uso
├── docker/              # Docker configs
│   ├── Dockerfile.api
│   ├── Dockerfile.cli
│   └── docker-compose.yml
└── scripts/             # Scripts auxiliares
```

## 🔐 Segurança

- **TLS 1.3**: Todas as comunicações criptografadas
- **mTLS**: Autenticação mútua com certificados
- **OAuth 2.0**: Padrão de autenticação
- **JWT**: Tokens com refresh automático
- **AES-256-GCM**: Criptografia de dados sensíveis
- **Rate Limiting**: Proteção contra abuso
- **CORS**: Configuração restritiva
- **SQL Injection**: Prepared statements
- **XSS Protection**: Sanitização de inputs

## 📊 Auditoria e Compliance

- **Logs Completos**: Todas as operações registradas
- **Retenção**: 5 anos (conforme legislação brasileira)
- **PCI DSS**: Compliance desde o design
- **LGPD**: Proteção de dados pessoais
- **Rastreabilidade**: Tracking completo de transações

## 🚀 Quick Start

### Pré-requisitos
- Go 1.21+
- PostgreSQL 15+
- Node.js 18+
- Docker & Docker Compose

### Instalação

```bash
# Clone o repositório
git clone <repo-url>
cd pix-saas

# Backend
cd backend
go mod download
go run cmd/api/main.go

# Frontend
cd ../frontend
npm install
npm run dev

# CLI Admin
cd ../backend
go run cmd/cli/main.go --help
```

## 📚 Documentação

- [Guia de Instalação](./docs/guides/installation.md)
- [Arquitetura Técnica](./docs/architecture/overview.md)
- [API Reference](./docs/api/openapi.yaml)
- [Guia de Integração](./docs/guides/integration.md)
- [Adicionar Novo Banco](./docs/guides/add-provider.md)

## 🔧 Configuração

### Variáveis de Ambiente

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=pixsaas
DB_USER=postgres
DB_PASSWORD=secret

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d

# Encryption
ENCRYPTION_KEY=your-32-byte-key

# Server
PORT=8080
ENV=development
```

## 🧪 Testes

```bash
# Backend
cd backend
go test ./... -v -cover

# Frontend
cd frontend
npm test
```

## 📦 Deploy

```bash
# Docker Compose
docker-compose up -d

# Kubernetes
kubectl apply -f k8s/
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📄 Licença

Proprietary - Todos os direitos reservados

## 📞 Suporte

- Email: suporte@pixsaas.com.br
- Documentação: https://docs.pixsaas.com.br
- Status: https://status.pixsaas.com.br

---

**Desenvolvido com ❤️ para o ecossistema financeiro brasileiro**
