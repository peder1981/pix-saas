#!/bin/bash

# Script de Autorun para Testes - PIX SaaS Platform
# Autor: Peder Munksgaard (JMPM Tecnologia)
# Data: 2025-01-19
# Descrição: Executa todos os testes automaticamente e corrige inconsistências

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para imprimir cabeçalho
print_header() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Função para imprimir status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
    fi
}

# Função para imprimir warning
print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Função para imprimir info
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Verificar se estamos no diretório correto
if [ ! -d "backend" ]; then
    echo -e "${RED}Erro: Execute este script da raiz do projeto${NC}"
    exit 1
fi

print_header "PIX SaaS Platform - Autorun de Testes"
print_info "Iniciando execução automática de todos os testes..."
echo ""

cd backend

# Contador de tentativas
ATTEMPT=1
MAX_ATTEMPTS=3
ALL_TESTS_PASSED=false

while [ $ATTEMPT -le $MAX_ATTEMPTS ] && [ "$ALL_TESTS_PASSED" = false ]; do
    print_header "Tentativa $ATTEMPT de $MAX_ATTEMPTS"
    
    # 1. Verificar e baixar dependências
    print_info "1. Verificando dependências..."
    go mod download 2>&1 | tee /tmp/go-mod-download.log
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        print_status 0 "Dependências baixadas"
    else
        print_status 1 "Erro ao baixar dependências"
        cat /tmp/go-mod-download.log
        exit 1
    fi
    
    go mod tidy 2>&1 | tee /tmp/go-mod-tidy.log
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        print_status 0 "Dependências organizadas"
    else
        print_warning "Aviso ao organizar dependências"
    fi
    echo ""
    
    # 2. Verificar compilação
    print_info "2. Verificando compilação..."
    go build -v ./... 2>&1 | tee /tmp/go-build.log
    BUILD_RESULT=${PIPESTATUS[0]}
    
    if [ $BUILD_RESULT -eq 0 ]; then
        print_status 0 "Compilação bem-sucedida"
    else
        print_status 1 "Erro na compilação"
        echo ""
        print_warning "Analisando erros de compilação..."
        
        # Verificar erros comuns
        if grep -q "undefined:" /tmp/go-build.log; then
            print_warning "Encontrados símbolos indefinidos"
            grep "undefined:" /tmp/go-build.log | head -5
        fi
        
        if grep -q "cannot use" /tmp/go-build.log; then
            print_warning "Encontrados erros de tipo"
            grep "cannot use" /tmp/go-build.log | head -5
        fi
        
        if [ $ATTEMPT -lt $MAX_ATTEMPTS ]; then
            print_info "Tentando corrigir automaticamente..."
            go get -u ./...
            go mod tidy
        else
            echo ""
            print_status 1 "Não foi possível corrigir automaticamente"
            exit 1
        fi
    fi
    echo ""
    
    # 3. Executar testes por pacote
    print_info "3. Executando testes por pacote..."
    echo ""
    
    # Array para armazenar resultados
    declare -a FAILED_PACKAGES=()
    declare -a PASSED_PACKAGES=()
    
    # Listar todos os pacotes com testes
    PACKAGES=$(go list ./... | grep -v /vendor/)
    
    for package in $PACKAGES; do
        # Verificar se o pacote tem testes
        if go list -f '{{if .TestGoFiles}}{{.ImportPath}}{{end}}' $package | grep -q .; then
            package_name=$(basename $package)
            echo -e "${BLUE}Testing:${NC} $package"
            
            # Executar testes do pacote
            if go test -v -count=1 $package 2>&1 | tee /tmp/test-$package_name.log; then
                PASSED_PACKAGES+=("$package")
                print_status 0 "$package_name"
            else
                FAILED_PACKAGES+=("$package")
                print_status 1 "$package_name"
                
                # Analisar falhas
                if grep -q "undefined:" /tmp/test-$package_name.log; then
                    print_warning "  → Símbolos indefinidos detectados"
                fi
                
                if grep -q "cannot use" /tmp/test-$package_name.log; then
                    print_warning "  → Erros de tipo detectados"
                fi
                
                if grep -q "FAIL" /tmp/test-$package_name.log; then
                    print_warning "  → Testes falharam"
                    grep "FAIL:" /tmp/test-$package_name.log | head -3
                fi
            fi
            echo ""
        fi
    done
    
    # 4. Resumo da tentativa
    print_header "Resumo da Tentativa $ATTEMPT"
    
    TOTAL_PACKAGES=$((${#PASSED_PACKAGES[@]} + ${#FAILED_PACKAGES[@]}))
    PASSED_COUNT=${#PASSED_PACKAGES[@]}
    FAILED_COUNT=${#FAILED_PACKAGES[@]}
    
    echo "Total de pacotes testados: $TOTAL_PACKAGES"
    echo -e "${GREEN}Pacotes OK: $PASSED_COUNT${NC}"
    echo -e "${RED}Pacotes com falha: $FAILED_COUNT${NC}"
    echo ""
    
    if [ $FAILED_COUNT -eq 0 ]; then
        ALL_TESTS_PASSED=true
        print_status 0 "TODOS OS TESTES PASSARAM!"
        break
    else
        print_warning "Pacotes com falha:"
        for pkg in "${FAILED_PACKAGES[@]}"; do
            echo "  - $pkg"
        done
        echo ""
        
        if [ $ATTEMPT -lt $MAX_ATTEMPTS ]; then
            print_info "Tentando corrigir e executar novamente..."
            sleep 2
        fi
    fi
    
    ATTEMPT=$((ATTEMPT + 1))
done

echo ""
print_header "Resultado Final"

if [ "$ALL_TESTS_PASSED" = true ]; then
    # 5. Gerar relatório de cobertura
    print_info "Gerando relatório de cobertura..."
    go test -coverprofile=coverage.out ./... > /dev/null 2>&1
    
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
        echo ""
        echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║                  ✓ SUCESSO TOTAL!                        ║${NC}"
        echo -e "${GREEN}╠═══════════════════════════════════════════════════════════╣${NC}"
        echo -e "${GREEN}║  Todos os testes passaram com sucesso!                   ║${NC}"
        echo -e "${GREEN}║  Cobertura de código: $COVERAGE                              ║${NC}"
        echo -e "${GREEN}║  Pacotes testados: $TOTAL_PACKAGES                                      ║${NC}"
        echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
        echo ""
        
        # Gerar HTML
        go tool cover -html=coverage.out -o coverage.html
        print_status 0 "Relatório HTML gerado: backend/coverage.html"
        
        # Estatísticas detalhadas
        echo ""
        print_info "Top 5 pacotes com melhor cobertura:"
        go tool cover -func=coverage.out | grep -v "total:" | sort -k3 -rn | head -5
        
    fi
    
    # 6. Compilar binários finais
    echo ""
    print_info "Compilando binários finais..."
    
    mkdir -p bin
    
    if go build -o bin/api ./cmd/api; then
        print_status 0 "API compilada: backend/bin/api"
    else
        print_status 1 "Erro ao compilar API"
    fi
    
    if go build -o bin/cli ./cmd/cli; then
        print_status 0 "CLI compilada: backend/bin/cli"
    else
        print_status 1 "Erro ao compilar CLI"
    fi
    
    # 7. Verificar tamanho dos binários
    echo ""
    print_info "Tamanho dos binários:"
    if [ -f bin/api ]; then
        API_SIZE=$(du -h bin/api | cut -f1)
        echo "  API: $API_SIZE"
    fi
    if [ -f bin/cli ]; then
        CLI_SIZE=$(du -h bin/cli | cut -f1)
        echo "  CLI: $CLI_SIZE"
    fi
    
    echo ""
    print_header "Próximos Passos"
    echo "1. Revisar relatório de cobertura: open backend/coverage.html"
    echo "2. Executar API: ./backend/bin/api"
    echo "3. Executar CLI: ./backend/bin/cli --help"
    echo "4. Fazer commit das mudanças"
    echo ""
    
    exit 0
else
    echo ""
    echo -e "${RED}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                  ✗ FALHA NOS TESTES                      ║${NC}"
    echo -e "${RED}╠═══════════════════════════════════════════════════════════╣${NC}"
    echo -e "${RED}║  Alguns testes falharam após $MAX_ATTEMPTS tentativas            ║${NC}"
    echo -e "${RED}║  Pacotes com falha: $FAILED_COUNT                                    ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    
    print_warning "Pacotes que falharam:"
    for pkg in "${FAILED_PACKAGES[@]}"; do
        echo "  - $pkg"
        pkg_name=$(basename $pkg)
        if [ -f /tmp/test-$pkg_name.log ]; then
            echo ""
            echo "    Últimas linhas do log:"
            tail -10 /tmp/test-$pkg_name.log | sed 's/^/    /'
            echo ""
        fi
    done
    
    echo ""
    print_info "Ações recomendadas:"
    echo "1. Revisar logs em /tmp/test-*.log"
    echo "2. Corrigir erros manualmente"
    echo "3. Executar novamente: ./scripts/autorun-tests.sh"
    echo ""
    
    exit 1
fi
