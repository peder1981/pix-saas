#!/bin/bash

# Script para validar CI localmente antes de push
# Simula os checks que serão executados no GitHub Actions

set -e

echo "🔍 Validando CI localmente..."
echo ""

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para imprimir com cor
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
        exit 1
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Verificar se estamos no diretório correto
if [ ! -d "backend" ]; then
    echo -e "${RED}Erro: Execute este script da raiz do projeto${NC}"
    exit 1
fi

cd backend

echo "📦 1. Verificando dependências..."
go mod download
go mod verify
print_status $? "Dependências verificadas"
echo ""

echo "🧪 2. Executando testes..."
go test -v -race -coverprofile=coverage.out ./...
TEST_RESULT=$?
print_status $TEST_RESULT "Testes executados"

if [ $TEST_RESULT -eq 0 ]; then
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "   Cobertura: ${COVERAGE}%"
    
    # Verificar se cobertura é maior que 70%
    if (( $(echo "$COVERAGE < 70" | bc -l) )); then
        print_warning "Cobertura abaixo de 70%"
    fi
fi
echo ""

echo "🔍 3. Executando linter..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
    print_status $? "Linter executado"
else
    print_warning "golangci-lint não instalado, pulando..."
    echo "   Instale com: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
fi
echo ""

echo "🔒 4. Verificando segurança..."
if command -v gosec &> /dev/null; then
    gosec -quiet ./...
    print_status $? "Scan de segurança executado"
else
    print_warning "gosec não instalado, pulando..."
    echo "   Instale com: go install github.com/securego/gosec/v2/cmd/gosec@latest"
fi
echo ""

echo "🏗️  5. Compilando binários..."
go build -o bin/api ./cmd/api
print_status $? "API compilada"

go build -o bin/cli ./cmd/cli
print_status $? "CLI compilada"
echo ""

echo "📝 6. Verificando formatação..."
UNFORMATTED=$(gofmt -l .)
if [ -z "$UNFORMATTED" ]; then
    print_status 0 "Código formatado corretamente"
else
    echo -e "${RED}✗${NC} Arquivos não formatados:"
    echo "$UNFORMATTED"
    echo ""
    echo "Execute: gofmt -w ."
    exit 1
fi
echo ""

echo "🔍 7. Verificando imports..."
if command -v goimports &> /dev/null; then
    UNFORMATTED=$(goimports -l .)
    if [ -z "$UNFORMATTED" ]; then
        print_status 0 "Imports organizados"
    else
        echo -e "${RED}✗${NC} Arquivos com imports desorganizados:"
        echo "$UNFORMATTED"
        echo ""
        echo "Execute: goimports -w ."
        exit 1
    fi
else
    print_warning "goimports não instalado, pulando..."
    echo "   Instale com: go install golang.org/x/tools/cmd/goimports@latest"
fi
echo ""

echo "🐳 8. Validando Dockerfile..."
cd ..
if command -v docker &> /dev/null; then
    docker build -f docker/Dockerfile.api -t pix-saas-api:test . > /dev/null 2>&1
    print_status $? "Dockerfile válido"
    docker rmi pix-saas-api:test > /dev/null 2>&1
else
    print_warning "Docker não instalado, pulando..."
fi
echo ""

echo "📋 9. Verificando arquivos obrigatórios..."
REQUIRED_FILES=(
    "README.md"
    "LICENSE"
    "CHANGELOG.md"
    ".gitignore"
    ".github/workflows/tests.yml"
    "backend/go.mod"
    "backend/go.sum"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "   ${GREEN}✓${NC} $file"
    else
        echo -e "   ${RED}✗${NC} $file (faltando)"
    fi
done
echo ""

echo "✨ Validação concluída!"
echo ""
echo "📊 Resumo:"
echo "   - Testes: OK"
echo "   - Cobertura: ${COVERAGE}%"
echo "   - Build: OK"
echo "   - Formatação: OK"
echo ""
echo "🚀 Pronto para push!"
