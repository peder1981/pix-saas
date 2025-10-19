#!/bin/bash

# PIX SaaS - Build Script (Linux/macOS)
# Autor: Peder Munksgaard
# Data: 19/10/2025

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_header "Verificando Pré-requisitos"
    
    # Check Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        print_success "Go instalado: $GO_VERSION"
    else
        print_error "Go não encontrado. Instale Go 1.21+ de https://golang.org"
        exit 1
    fi
    
    # Check Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | awk '{print $3}')
        print_success "Docker instalado: $DOCKER_VERSION"
    else
        print_warning "Docker não encontrado. Algumas funcionalidades não estarão disponíveis."
    fi
    
    # Check PostgreSQL
    if command -v psql &> /dev/null; then
        PSQL_VERSION=$(psql --version | awk '{print $3}')
        print_success "PostgreSQL instalado: $PSQL_VERSION"
    else
        print_warning "PostgreSQL não encontrado. Use Docker ou instale PostgreSQL 15+"
    fi
    
    # Check Node.js
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        print_success "Node.js instalado: $NODE_VERSION"
    else
        print_warning "Node.js não encontrado. Frontend não será compilado."
    fi
    
    echo ""
}

# Setup environment
setup_environment() {
    print_header "Configurando Ambiente"
    
    # Create .env if not exists
    if [ ! -f "backend/.env" ]; then
        print_info "Criando arquivo .env..."
        cp backend/.env.example backend/.env
        
        # Generate keys
        ENCRYPTION_KEY=$(openssl rand -base64 32)
        JWT_SECRET=$(openssl rand -base64 64)
        
        # Update .env
        sed -i.bak "s|your-base64-encoded-32-byte-encryption-key|$ENCRYPTION_KEY|g" backend/.env
        sed -i.bak "s|your-super-secret-jwt-key-change-this-in-production|$JWT_SECRET|g" backend/.env
        rm backend/.env.bak
        
        print_success "Arquivo .env criado com chaves geradas"
    else
        print_success "Arquivo .env já existe"
    fi
    
    echo ""
}

# Build backend
build_backend() {
    print_header "Compilando Backend"
    
    cd backend
    
    # Download dependencies
    print_info "Baixando dependências..."
    go mod download
    print_success "Dependências baixadas"
    
    # Build API
    print_info "Compilando API..."
    go build -o ../bin/api cmd/api/main.go
    print_success "API compilada: bin/api"
    
    # Build CLI
    print_info "Compilando CLI..."
    go build -o ../bin/pixsaas-cli cmd/cli/main.go
    print_success "CLI compilada: bin/pixsaas-cli"
    
    cd ..
    echo ""
}

# Build frontend
build_frontend() {
    print_header "Compilando Frontend"
    
    if ! command -v node &> /dev/null; then
        print_warning "Node.js não encontrado. Pulando frontend."
        return
    fi
    
    cd frontend
    
    # Install dependencies
    print_info "Instalando dependências..."
    npm install
    print_success "Dependências instaladas"
    
    # Build
    print_info "Compilando frontend..."
    npm run build
    print_success "Frontend compilado"
    
    cd ..
    echo ""
}

# Setup database
setup_database() {
    print_header "Configurando Banco de Dados"
    
    # Check if PostgreSQL is running
    if ! command -v psql &> /dev/null; then
        print_warning "PostgreSQL não encontrado. Use Docker: docker-compose up -d postgres"
        return
    fi
    
    # Database name from .env or default
    DB_NAME=${DB_NAME:-pixsaas}
    
    # Check if database exists
    if psql -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
        print_success "Banco de dados '$DB_NAME' já existe"
    else
        print_info "Criando banco de dados '$DB_NAME'..."
        createdb $DB_NAME
        print_success "Banco de dados criado"
    fi
    
    # Run migrations
    print_info "Executando migrations..."
    psql -d $DB_NAME -f backend/migrations/001_initial_schema.sql > /dev/null 2>&1
    print_success "Migrations executadas"
    
    echo ""
}

# Build Docker images
build_docker() {
    print_header "Construindo Imagens Docker"
    
    if ! command -v docker &> /dev/null; then
        print_warning "Docker não encontrado. Pulando build de imagens."
        return
    fi
    
    print_info "Construindo imagem da API..."
    docker build -f docker/Dockerfile.api -t pixsaas-api:latest .
    print_success "Imagem pixsaas-api:latest criada"
    
    echo ""
}

# Run tests
run_tests() {
    print_header "Executando Testes"
    
    cd backend
    
    print_info "Executando testes..."
    go test ./... -v
    
    if [ $? -eq 0 ]; then
        print_success "Todos os testes passaram"
    else
        print_error "Alguns testes falharam"
    fi
    
    cd ..
    echo ""
}

# Main menu
show_menu() {
    echo ""
    print_header "PIX SaaS - Build Script"
    echo "Escolha uma opção:"
    echo ""
    echo "  1) Build Completo (Backend + Frontend + Docker)"
    echo "  2) Build Backend"
    echo "  3) Build Frontend"
    echo "  4) Build Docker"
    echo "  5) Setup Banco de Dados"
    echo "  6) Executar Testes"
    echo "  7) Iniciar com Docker Compose"
    echo "  8) Limpar builds"
    echo "  9) Sair"
    echo ""
    read -p "Opção: " choice
    echo ""
}

# Clean builds
clean_builds() {
    print_header "Limpando Builds"
    
    print_info "Removendo binários..."
    rm -rf bin/
    print_success "Binários removidos"
    
    print_info "Limpando cache Go..."
    cd backend && go clean -cache && cd ..
    print_success "Cache Go limpo"
    
    if [ -d "frontend/.next" ]; then
        print_info "Removendo build do frontend..."
        rm -rf frontend/.next
        print_success "Build do frontend removido"
    fi
    
    echo ""
}

# Start with Docker Compose
start_docker_compose() {
    print_header "Iniciando com Docker Compose"
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose não encontrado"
        return
    fi
    
    print_info "Iniciando containers..."
    docker-compose up -d
    
    echo ""
    print_success "Containers iniciados!"
    echo ""
    print_info "Serviços disponíveis:"
    echo "  - API: http://localhost:8080"
    echo "  - PostgreSQL: localhost:5432"
    echo ""
    print_info "Para ver logs: docker-compose logs -f"
    print_info "Para parar: docker-compose down"
    echo ""
}

# Main execution
main() {
    # Check if we're in the right directory
    if [ ! -f "README.md" ]; then
        print_error "Execute este script na raiz do projeto PIX SaaS"
        exit 1
    fi
    
    # Create bin directory
    mkdir -p bin
    
    # Interactive mode
    if [ "$1" == "" ]; then
        while true; do
            show_menu
            case $choice in
                1)
                    check_prerequisites
                    setup_environment
                    build_backend
                    build_frontend
                    build_docker
                    print_success "Build completo finalizado!"
                    ;;
                2)
                    check_prerequisites
                    setup_environment
                    build_backend
                    ;;
                3)
                    build_frontend
                    ;;
                4)
                    build_docker
                    ;;
                5)
                    setup_database
                    ;;
                6)
                    run_tests
                    ;;
                7)
                    start_docker_compose
                    ;;
                8)
                    clean_builds
                    ;;
                9)
                    print_info "Saindo..."
                    exit 0
                    ;;
                *)
                    print_error "Opção inválida"
                    ;;
            esac
            
            read -p "Pressione ENTER para continuar..."
        done
    fi
    
    # Command line mode
    case "$1" in
        all)
            check_prerequisites
            setup_environment
            build_backend
            build_frontend
            build_docker
            ;;
        backend)
            check_prerequisites
            setup_environment
            build_backend
            ;;
        frontend)
            build_frontend
            ;;
        docker)
            build_docker
            ;;
        db)
            setup_database
            ;;
        test)
            run_tests
            ;;
        start)
            start_docker_compose
            ;;
        clean)
            clean_builds
            ;;
        *)
            echo "Uso: $0 [all|backend|frontend|docker|db|test|start|clean]"
            echo ""
            echo "Ou execute sem argumentos para modo interativo"
            exit 1
            ;;
    esac
}

# Run main
main "$@"
