# PIX SaaS - Build Script (PowerShell/Windows)
# Autor: Peder Munksgaard
# Data: 19/10/2025

# Requires PowerShell 5.1 or higher

param(
    [Parameter(Position=0)]
    [ValidateSet('all', 'backend', 'frontend', 'docker', 'db', 'test', 'start', 'clean', '')]
    [string]$Action = ''
)

# Colors
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = 'White'
    )
    Write-Host $Message -ForegroundColor $Color
}

function Print-Header {
    param([string]$Title)
    Write-ColorOutput "`n================================" -Color Cyan
    Write-ColorOutput $Title -Color Cyan
    Write-ColorOutput "================================`n" -Color Cyan
}

function Print-Success {
    param([string]$Message)
    Write-ColorOutput "✓ $Message" -Color Green
}

function Print-Error {
    param([string]$Message)
    Write-ColorOutput "✗ $Message" -Color Red
}

function Print-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠ $Message" -Color Yellow
}

function Print-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ $Message" -Color Cyan
}

# Check prerequisites
function Check-Prerequisites {
    Print-Header "Verificando Pré-requisitos"
    
    # Check Go
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $goVersion = (go version).Split()[2]
        Print-Success "Go instalado: $goVersion"
    } else {
        Print-Error "Go não encontrado. Instale Go 1.21+ de https://golang.org"
        exit 1
    }
    
    # Check Docker
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        $dockerVersion = (docker --version).Split()[2]
        Print-Success "Docker instalado: $dockerVersion"
    } else {
        Print-Warning "Docker não encontrado. Algumas funcionalidades não estarão disponíveis."
    }
    
    # Check PostgreSQL
    if (Get-Command psql -ErrorAction SilentlyContinue) {
        $psqlVersion = (psql --version).Split()[2]
        Print-Success "PostgreSQL instalado: $psqlVersion"
    } else {
        Print-Warning "PostgreSQL não encontrado. Use Docker ou instale PostgreSQL 15+"
    }
    
    # Check Node.js
    if (Get-Command node -ErrorAction SilentlyContinue) {
        $nodeVersion = node --version
        Print-Success "Node.js instalado: $nodeVersion"
    } else {
        Print-Warning "Node.js não encontrado. Frontend não será compilado."
    }
    
    Write-Host ""
}

# Setup environment
function Setup-Environment {
    Print-Header "Configurando Ambiente"
    
    # Create .env if not exists
    if (-not (Test-Path "backend\.env")) {
        Print-Info "Criando arquivo .env..."
        Copy-Item "backend\.env.example" "backend\.env"
        
        # Generate keys (requires OpenSSL or use .NET)
        $encryptionKey = [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
        $jwtSecret = [Convert]::ToBase64String((1..64 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))
        
        # Update .env
        $envContent = Get-Content "backend\.env"
        $envContent = $envContent -replace 'your-base64-encoded-32-byte-encryption-key', $encryptionKey
        $envContent = $envContent -replace 'your-super-secret-jwt-key-change-this-in-production', $jwtSecret
        $envContent | Set-Content "backend\.env"
        
        Print-Success "Arquivo .env criado com chaves geradas"
    } else {
        Print-Success "Arquivo .env já existe"
    }
    
    Write-Host ""
}

# Build backend
function Build-Backend {
    Print-Header "Compilando Backend"
    
    Push-Location backend
    
    # Download dependencies
    Print-Info "Baixando dependências..."
    go mod download
    Print-Success "Dependências baixadas"
    
    # Create bin directory
    if (-not (Test-Path "..\bin")) {
        New-Item -ItemType Directory -Path "..\bin" | Out-Null
    }
    
    # Build API
    Print-Info "Compilando API..."
    go build -o ..\bin\api.exe cmd\api\main.go
    Print-Success "API compilada: bin\api.exe"
    
    # Build CLI
    Print-Info "Compilando CLI..."
    go build -o ..\bin\pixsaas-cli.exe cmd\cli\main.go
    Print-Success "CLI compilada: bin\pixsaas-cli.exe"
    
    Pop-Location
    Write-Host ""
}

# Build frontend
function Build-Frontend {
    Print-Header "Compilando Frontend"
    
    if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
        Print-Warning "Node.js não encontrado. Pulando frontend."
        return
    }
    
    Push-Location frontend
    
    # Install dependencies
    Print-Info "Instalando dependências..."
    npm install
    Print-Success "Dependências instaladas"
    
    # Build
    Print-Info "Compilando frontend..."
    npm run build
    Print-Success "Frontend compilado"
    
    Pop-Location
    Write-Host ""
}

# Setup database
function Setup-Database {
    Print-Header "Configurando Banco de Dados"
    
    # Check if PostgreSQL is running
    if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
        Print-Warning "PostgreSQL não encontrado. Use Docker: docker-compose up -d postgres"
        return
    }
    
    # Database name from .env or default
    $dbName = "pixsaas"
    
    # Check if database exists
    $dbExists = psql -lqt | Select-String -Pattern $dbName
    
    if ($dbExists) {
        Print-Success "Banco de dados '$dbName' já existe"
    } else {
        Print-Info "Criando banco de dados '$dbName'..."
        createdb $dbName
        Print-Success "Banco de dados criado"
    }
    
    # Run migrations
    Print-Info "Executando migrations..."
    psql -d $dbName -f backend\migrations\001_initial_schema.sql | Out-Null
    Print-Success "Migrations executadas"
    
    Write-Host ""
}

# Build Docker images
function Build-Docker {
    Print-Header "Construindo Imagens Docker"
    
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Print-Warning "Docker não encontrado. Pulando build de imagens."
        return
    }
    
    Print-Info "Construindo imagem da API..."
    docker build -f docker\Dockerfile.api -t pixsaas-api:latest .
    Print-Success "Imagem pixsaas-api:latest criada"
    
    Write-Host ""
}

# Run tests
function Run-Tests {
    Print-Header "Executando Testes"
    
    Push-Location backend
    
    Print-Info "Executando testes..."
    go test .\... -v
    
    if ($LASTEXITCODE -eq 0) {
        Print-Success "Todos os testes passaram"
    } else {
        Print-Error "Alguns testes falharam"
    }
    
    Pop-Location
    Write-Host ""
}

# Clean builds
function Clean-Builds {
    Print-Header "Limpando Builds"
    
    Print-Info "Removendo binários..."
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force bin
    }
    Print-Success "Binários removidos"
    
    Print-Info "Limpando cache Go..."
    Push-Location backend
    go clean -cache
    Pop-Location
    Print-Success "Cache Go limpo"
    
    if (Test-Path "frontend\.next") {
        Print-Info "Removendo build do frontend..."
        Remove-Item -Recurse -Force frontend\.next
        Print-Success "Build do frontend removido"
    }
    
    Write-Host ""
}

# Start with Docker Compose
function Start-DockerCompose {
    Print-Header "Iniciando com Docker Compose"
    
    if (-not (Get-Command docker-compose -ErrorAction SilentlyContinue)) {
        Print-Error "Docker Compose não encontrado"
        return
    }
    
    Print-Info "Iniciando containers..."
    docker-compose up -d
    
    Write-Host ""
    Print-Success "Containers iniciados!"
    Write-Host ""
    Print-Info "Serviços disponíveis:"
    Write-Host "  - API: http://localhost:8080"
    Write-Host "  - PostgreSQL: localhost:5432"
    Write-Host ""
    Print-Info "Para ver logs: docker-compose logs -f"
    Print-Info "Para parar: docker-compose down"
    Write-Host ""
}

# Show menu
function Show-Menu {
    Write-Host ""
    Print-Header "PIX SaaS - Build Script"
    Write-Host "Escolha uma opção:"
    Write-Host ""
    Write-Host "  1) Build Completo (Backend + Frontend + Docker)"
    Write-Host "  2) Build Backend"
    Write-Host "  3) Build Frontend"
    Write-Host "  4) Build Docker"
    Write-Host "  5) Setup Banco de Dados"
    Write-Host "  6) Executar Testes"
    Write-Host "  7) Iniciar com Docker Compose"
    Write-Host "  8) Limpar builds"
    Write-Host "  9) Sair"
    Write-Host ""
    $choice = Read-Host "Opção"
    return $choice
}

# Main execution
function Main {
    # Check if we're in the right directory
    if (-not (Test-Path "README.md")) {
        Print-Error "Execute este script na raiz do projeto PIX SaaS"
        exit 1
    }
    
    # Create bin directory
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    
    # Interactive mode
    if ($Action -eq '') {
        while ($true) {
            $choice = Show-Menu
            Write-Host ""
            
            switch ($choice) {
                '1' {
                    Check-Prerequisites
                    Setup-Environment
                    Build-Backend
                    Build-Frontend
                    Build-Docker
                    Print-Success "Build completo finalizado!"
                }
                '2' {
                    Check-Prerequisites
                    Setup-Environment
                    Build-Backend
                }
                '3' {
                    Build-Frontend
                }
                '4' {
                    Build-Docker
                }
                '5' {
                    Setup-Database
                }
                '6' {
                    Run-Tests
                }
                '7' {
                    Start-DockerCompose
                }
                '8' {
                    Clean-Builds
                }
                '9' {
                    Print-Info "Saindo..."
                    exit 0
                }
                default {
                    Print-Error "Opção inválida"
                }
            }
            
            Read-Host "`nPressione ENTER para continuar..."
        }
    }
    
    # Command line mode
    switch ($Action) {
        'all' {
            Check-Prerequisites
            Setup-Environment
            Build-Backend
            Build-Frontend
            Build-Docker
        }
        'backend' {
            Check-Prerequisites
            Setup-Environment
            Build-Backend
        }
        'frontend' {
            Build-Frontend
        }
        'docker' {
            Build-Docker
        }
        'db' {
            Setup-Database
        }
        'test' {
            Run-Tests
        }
        'start' {
            Start-DockerCompose
        }
        'clean' {
            Clean-Builds
        }
        default {
            Write-Host "Uso: .\build.ps1 [all|backend|frontend|docker|db|test|start|clean]"
            Write-Host ""
            Write-Host "Ou execute sem argumentos para modo interativo"
            exit 1
        }
    }
}

# Run main
Main
