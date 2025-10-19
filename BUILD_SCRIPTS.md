# Scripts de Build - PIX SaaS

Este documento descreve como usar os scripts de build para compilar e executar o projeto PIX SaaS em diferentes sistemas operacionais.

## 📋 Scripts Disponíveis

- **`build.sh`** - Script Bash para Linux/macOS
- **`build.ps1`** - Script PowerShell para Windows
- **`build.bat`** - Script Batch para Windows (CMD)

## 🚀 Uso Rápido

### Linux/macOS (Bash)

```bash
# Modo interativo (menu)
./build.sh

# Build completo
./build.sh all

# Build apenas backend
./build.sh backend

# Build apenas frontend
./build.sh frontend

# Build Docker images
./build.sh docker

# Setup banco de dados
./build.sh db

# Executar testes
./build.sh test

# Iniciar com Docker Compose
./build.sh start

# Limpar builds
./build.sh clean
```

### Windows (PowerShell)

```powershell
# Modo interativo (menu)
.\build.ps1

# Build completo
.\build.ps1 all

# Build apenas backend
.\build.ps1 backend

# Build apenas frontend
.\build.ps1 frontend

# Build Docker images
.\build.ps1 docker

# Setup banco de dados
.\build.ps1 db

# Executar testes
.\build.ps1 test

# Iniciar com Docker Compose
.\build.ps1 start

# Limpar builds
.\build.ps1 clean
```

### Windows (CMD/Batch)

```batch
REM Modo interativo (menu)
build.bat

REM Escolha a opção no menu:
REM 1 - Build Completo
REM 2 - Build Backend
REM 3 - Build Frontend
REM 4 - Build Docker
REM 5 - Setup Banco de Dados
REM 6 - Executar Testes
REM 7 - Iniciar com Docker Compose
REM 8 - Limpar builds
REM 9 - Sair
```

## 📦 O Que Cada Opção Faz

### 1. Build Completo
- Verifica pré-requisitos
- Configura ambiente (.env)
- Compila backend (API + CLI)
- Compila frontend
- Constrói imagens Docker

**Resultado**:
- `bin/api` ou `bin/api.exe` - Servidor API
- `bin/pixsaas-cli` ou `bin/pixsaas-cli.exe` - CLI administrativa
- `frontend/.next/` - Frontend compilado
- Imagem Docker `pixsaas-api:latest`

### 2. Build Backend
- Baixa dependências Go
- Compila API server
- Compila CLI administrativa

**Resultado**:
- `bin/api` ou `bin/api.exe`
- `bin/pixsaas-cli` ou `bin/pixsaas-cli.exe`

### 3. Build Frontend
- Instala dependências npm
- Compila aplicação Next.js

**Resultado**:
- `frontend/.next/` - Build otimizado
- `frontend/out/` - Export estático (se configurado)

### 4. Build Docker
- Constrói imagem Docker da API
- Usa Dockerfile multi-stage

**Resultado**:
- Imagem `pixsaas-api:latest`

### 5. Setup Banco de Dados
- Cria banco de dados PostgreSQL
- Executa migrations
- Configura schema inicial

**Resultado**:
- Banco `pixsaas` criado
- 10 tabelas criadas
- Índices configurados

### 6. Executar Testes
- Executa testes unitários Go
- Mostra cobertura

**Resultado**:
- Relatório de testes
- Status de aprovação/falha

### 7. Iniciar com Docker Compose
- Inicia todos os containers
- PostgreSQL + API + Monitoring (opcional)

**Resultado**:
- API em http://localhost:8080
- PostgreSQL em localhost:5432
- Prometheus em http://localhost:9090 (opcional)
- Grafana em http://localhost:3001 (opcional)

### 8. Limpar Builds
- Remove binários compilados
- Limpa cache Go
- Remove builds do frontend

**Resultado**:
- Diretório `bin/` removido
- Cache limpo
- Espaço em disco liberado

## 🔧 Pré-requisitos

### Obrigatórios
- **Go 1.21+** - https://golang.org
- **Git** - Para clonar o repositório

### Opcionais (mas recomendados)
- **Docker** - https://docker.com
- **Docker Compose** - Incluído no Docker Desktop
- **PostgreSQL 15+** - https://postgresql.org
- **Node.js 18+** - https://nodejs.org (para frontend)

## 📝 Configuração Inicial

### 1. Clonar Repositório
```bash
git clone https://github.com/peder1981/pix-saas.git
cd pix-saas
```

### 2. Executar Script de Build
```bash
# Linux/macOS
./build.sh

# Windows PowerShell
.\build.ps1

# Windows CMD
build.bat
```

### 3. Escolher Opção 1 (Build Completo)
O script irá:
- Verificar pré-requisitos
- Criar arquivo .env com chaves geradas
- Compilar backend
- Compilar frontend (se Node.js disponível)
- Construir imagens Docker (se Docker disponível)

### 4. Configurar Variáveis de Ambiente
Edite `backend/.env` e configure:
- `JWT_SECRET_KEY` - Chave JWT (gerada automaticamente)
- `ENCRYPTION_KEY` - Chave de criptografia (gerada automaticamente)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` - Configurações do banco

### 5. Iniciar Aplicação

#### Opção A: Com Docker Compose (Recomendado)
```bash
# Linux/macOS
./build.sh start

# Windows PowerShell
.\build.ps1 start

# Windows CMD
build.bat
# Escolha opção 7
```

#### Opção B: Manualmente
```bash
# 1. Iniciar PostgreSQL
# (ou usar Docker: docker-compose up -d postgres)

# 2. Executar migrations
./build.sh db

# 3. Iniciar API
./bin/api  # Linux/macOS
.\bin\api.exe  # Windows

# 4. Iniciar Frontend (em outro terminal)
cd frontend
npm run dev
```

## 🧪 Testando a Instalação

### 1. Verificar API
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

### 2. Testar CLI
```bash
# Linux/macOS
./bin/pixsaas-cli provider list

# Windows
.\bin\pixsaas-cli.exe provider list
```

### 3. Acessar Frontend
Abra o navegador em: http://localhost:3000

## 🐛 Troubleshooting

### Erro: "Go não encontrado"
**Solução**: Instale Go de https://golang.org

### Erro: "Permission denied" (Linux/macOS)
**Solução**: 
```bash
chmod +x build.sh
./build.sh
```

### Erro: "Execution Policy" (PowerShell)
**Solução**:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\build.ps1
```

### Erro: "Database connection failed"
**Solução**:
1. Verifique se PostgreSQL está rodando
2. Verifique credenciais em `backend/.env`
3. Ou use Docker: `docker-compose up -d postgres`

### Erro: "Port 8080 already in use"
**Solução**:
1. Pare o processo usando a porta
2. Ou mude a porta em `backend/.env`: `PORT=8081`

### Erro: "go.mod not found"
**Solução**: Execute o script na raiz do projeto

## 📊 Estrutura de Builds

```
pix-saas/
├── bin/                    # Binários compilados
│   ├── api                # API server (Linux/macOS)
│   ├── api.exe            # API server (Windows)
│   ├── pixsaas-cli        # CLI (Linux/macOS)
│   └── pixsaas-cli.exe    # CLI (Windows)
├── frontend/.next/        # Frontend compilado
├── backend/.env           # Configurações (gerado)
└── docker images          # Imagens Docker
    └── pixsaas-api:latest
```

## 🔄 Workflow Recomendado

### Desenvolvimento
```bash
# 1. Build inicial
./build.sh all

# 2. Durante desenvolvimento
./build.sh backend  # Recompilar backend
./build.sh frontend # Recompilar frontend

# 3. Testar mudanças
./build.sh test

# 4. Limpar e rebuild
./build.sh clean
./build.sh all
```

### Produção
```bash
# 1. Build completo
./build.sh all

# 2. Executar testes
./build.sh test

# 3. Build Docker
./build.sh docker

# 4. Deploy com Docker Compose
./build.sh start
```

## 📚 Comandos Úteis

### Ver logs da API
```bash
# Docker
docker-compose logs -f api

# Binário direto
./bin/api  # Logs no terminal
```

### Parar containers
```bash
docker-compose down
```

### Rebuild completo
```bash
./build.sh clean
./build.sh all
```

### Atualizar dependências
```bash
cd backend
go mod tidy
go mod download
```

## 🎯 Próximos Passos

Após build bem-sucedido:

1. ✅ Leia [INSTALL.md](./INSTALL.md) para configuração detalhada
2. ✅ Leia [README.md](./README.md) para visão geral
3. ✅ Configure credenciais bancárias
4. ✅ Teste endpoints da API
5. ✅ Configure webhooks
6. ✅ Deploy em produção

## 💡 Dicas

- Use **modo interativo** para facilitar: `./build.sh` sem argumentos
- Execute **build completo** na primeira vez
- Use **Docker Compose** para ambiente consistente
- Mantenha **.env** seguro e nunca commite
- Execute **testes** antes de deploy
- Use **clean** se tiver problemas de cache

## 📞 Suporte

Se encontrar problemas:

1. Verifique pré-requisitos
2. Leia mensagens de erro
3. Consulte [INSTALL.md](./INSTALL.md)
4. Abra issue no GitHub

---

**Desenvolvido para facilitar o desenvolvimento do PIX SaaS! 🚀**
