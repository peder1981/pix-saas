@echo off
REM PIX SaaS - Build Script (Batch/Windows)
REM Autor: Peder Munksgaard
REM Data: 19/10/2025

setlocal enabledelayedexpansion

REM Colors (limited in batch)
set "GREEN=[92m"
set "RED=[91m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM Main menu
:MENU
cls
echo.
echo %BLUE%================================%NC%
echo %BLUE%PIX SaaS - Build Script%NC%
echo %BLUE%================================%NC%
echo.
echo Escolha uma opcao:
echo.
echo   1) Build Completo (Backend + Frontend + Docker)
echo   2) Build Backend
echo   3) Build Frontend
echo   4) Build Docker
echo   5) Setup Banco de Dados
echo   6) Executar Testes
echo   7) Iniciar com Docker Compose
echo   8) Limpar builds
echo   9) Sair
echo.
set /p choice="Opcao: "
echo.

if "%choice%"=="1" goto BUILD_ALL
if "%choice%"=="2" goto BUILD_BACKEND
if "%choice%"=="3" goto BUILD_FRONTEND
if "%choice%"=="4" goto BUILD_DOCKER
if "%choice%"=="5" goto SETUP_DB
if "%choice%"=="6" goto RUN_TESTS
if "%choice%"=="7" goto START_DOCKER
if "%choice%"=="8" goto CLEAN_BUILDS
if "%choice%"=="9" goto END
goto INVALID_OPTION

REM Check prerequisites
:CHECK_PREREQ
echo.
echo %BLUE%================================%NC%
echo %BLUE%Verificando Pre-requisitos%NC%
echo %BLUE%================================%NC%
echo.

REM Check Go
where go >nul 2>nul
if %errorlevel% equ 0 (
    for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
    echo %GREEN%[OK] Go instalado: !GO_VERSION!%NC%
) else (
    echo %RED%[ERRO] Go nao encontrado. Instale Go 1.21+ de https://golang.org%NC%
    pause
    exit /b 1
)

REM Check Docker
where docker >nul 2>nul
if %errorlevel% equ 0 (
    echo %GREEN%[OK] Docker instalado%NC%
) else (
    echo %YELLOW%[AVISO] Docker nao encontrado%NC%
)

REM Check PostgreSQL
where psql >nul 2>nul
if %errorlevel% equ 0 (
    echo %GREEN%[OK] PostgreSQL instalado%NC%
) else (
    echo %YELLOW%[AVISO] PostgreSQL nao encontrado%NC%
)

REM Check Node.js
where node >nul 2>nul
if %errorlevel% equ 0 (
    for /f "tokens=*" %%i in ('node --version') do set NODE_VERSION=%%i
    echo %GREEN%[OK] Node.js instalado: !NODE_VERSION!%NC%
) else (
    echo %YELLOW%[AVISO] Node.js nao encontrado%NC%
)

echo.
goto :eof

REM Setup environment
:SETUP_ENV
echo.
echo %BLUE%================================%NC%
echo %BLUE%Configurando Ambiente%NC%
echo %BLUE%================================%NC%
echo.

if not exist "backend\.env" (
    echo Criando arquivo .env...
    copy "backend\.env.example" "backend\.env" >nul
    echo %GREEN%[OK] Arquivo .env criado%NC%
    echo %YELLOW%[AVISO] Edite backend\.env e adicione suas chaves%NC%
) else (
    echo %GREEN%[OK] Arquivo .env ja existe%NC%
)

echo.
goto :eof

REM Build backend
:BUILD_BACKEND_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Compilando Backend%NC%
echo %BLUE%================================%NC%
echo.

cd backend

echo Baixando dependencias...
go mod download
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Falha ao baixar dependencias%NC%
    cd ..
    goto :eof
)
echo %GREEN%[OK] Dependencias baixadas%NC%

REM Create bin directory
if not exist "..\bin" mkdir "..\bin"

echo Compilando API...
go build -o ..\bin\api.exe cmd\api\main.go
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Falha ao compilar API%NC%
    cd ..
    goto :eof
)
echo %GREEN%[OK] API compilada: bin\api.exe%NC%

echo Compilando CLI...
go build -o ..\bin\pixsaas-cli.exe cmd\cli\main.go
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Falha ao compilar CLI%NC%
    cd ..
    goto :eof
)
echo %GREEN%[OK] CLI compilada: bin\pixsaas-cli.exe%NC%

cd ..
echo.
goto :eof

REM Build frontend
:BUILD_FRONTEND_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Compilando Frontend%NC%
echo %BLUE%================================%NC%
echo.

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo %YELLOW%[AVISO] Node.js nao encontrado. Pulando frontend.%NC%
    goto :eof
)

cd frontend

echo Instalando dependencias...
call npm install
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Falha ao instalar dependencias%NC%
    cd ..
    goto :eof
)
echo %GREEN%[OK] Dependencias instaladas%NC%

echo Compilando frontend...
call npm run build
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Falha ao compilar frontend%NC%
    cd ..
    goto :eof
)
echo %GREEN%[OK] Frontend compilado%NC%

cd ..
echo.
goto :eof

REM Build Docker
:BUILD_DOCKER_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Construindo Imagens Docker%NC%
echo %BLUE%================================%NC%
echo.

where docker >nul 2>nul
if %errorlevel% neq 0 (
    echo %YELLOW%[AVISO] Docker nao encontrado. Pulando build de imagens.%NC%
    goto :eof
)

echo Construindo imagem da API...
docker build -f docker\Dockerfile.api -t pixsaas-api:latest .
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Falha ao construir imagem%NC%
    goto :eof
)
echo %GREEN%[OK] Imagem pixsaas-api:latest criada%NC%

echo.
goto :eof

REM Setup database
:SETUP_DB_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Configurando Banco de Dados%NC%
echo %BLUE%================================%NC%
echo.

where psql >nul 2>nul
if %errorlevel% neq 0 (
    echo %YELLOW%[AVISO] PostgreSQL nao encontrado. Use Docker: docker-compose up -d postgres%NC%
    goto :eof
)

set DB_NAME=pixsaas

echo Verificando banco de dados...
psql -lqt | findstr /C:"%DB_NAME%" >nul
if %errorlevel% equ 0 (
    echo %GREEN%[OK] Banco de dados '%DB_NAME%' ja existe%NC%
) else (
    echo Criando banco de dados '%DB_NAME%'...
    createdb %DB_NAME%
    echo %GREEN%[OK] Banco de dados criado%NC%
)

echo Executando migrations...
psql -d %DB_NAME% -f backend\migrations\001_initial_schema.sql >nul 2>nul
echo %GREEN%[OK] Migrations executadas%NC%

echo.
goto :eof

REM Run tests
:RUN_TESTS_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Executando Testes%NC%
echo %BLUE%================================%NC%
echo.

cd backend
go test .\... -v
cd ..

echo.
goto :eof

REM Clean builds
:CLEAN_BUILDS_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Limpando Builds%NC%
echo %BLUE%================================%NC%
echo.

if exist "bin" (
    echo Removendo binarios...
    rmdir /s /q bin
    echo %GREEN%[OK] Binarios removidos%NC%
)

echo Limpando cache Go...
cd backend
go clean -cache
cd ..
echo %GREEN%[OK] Cache Go limpo%NC%

if exist "frontend\.next" (
    echo Removendo build do frontend...
    rmdir /s /q frontend\.next
    echo %GREEN%[OK] Build do frontend removido%NC%
)

echo.
goto :eof

REM Start Docker Compose
:START_DOCKER_FUNC
echo.
echo %BLUE%================================%NC%
echo %BLUE%Iniciando com Docker Compose%NC%
echo %BLUE%================================%NC%
echo.

where docker-compose >nul 2>nul
if %errorlevel% neq 0 (
    echo %RED%[ERRO] Docker Compose nao encontrado%NC%
    goto :eof
)

echo Iniciando containers...
docker-compose up -d

echo.
echo %GREEN%[OK] Containers iniciados!%NC%
echo.
echo Servicos disponiveis:
echo   - API: http://localhost:8080
echo   - PostgreSQL: localhost:5432
echo.
echo Para ver logs: docker-compose logs -f
echo Para parar: docker-compose down
echo.
goto :eof

REM Menu actions
:BUILD_ALL
call :CHECK_PREREQ
call :SETUP_ENV
call :BUILD_BACKEND_FUNC
call :BUILD_FRONTEND_FUNC
call :BUILD_DOCKER_FUNC
echo %GREEN%[OK] Build completo finalizado!%NC%
pause
goto MENU

:BUILD_BACKEND
call :CHECK_PREREQ
call :SETUP_ENV
call :BUILD_BACKEND_FUNC
pause
goto MENU

:BUILD_FRONTEND
call :BUILD_FRONTEND_FUNC
pause
goto MENU

:BUILD_DOCKER
call :BUILD_DOCKER_FUNC
pause
goto MENU

:SETUP_DB
call :SETUP_DB_FUNC
pause
goto MENU

:RUN_TESTS
call :RUN_TESTS_FUNC
pause
goto MENU

:START_DOCKER
call :START_DOCKER_FUNC
pause
goto MENU

:CLEAN_BUILDS
call :CLEAN_BUILDS_FUNC
pause
goto MENU

:INVALID_OPTION
echo %RED%[ERRO] Opcao invalida%NC%
pause
goto MENU

:END
echo.
echo Saindo...
exit /b 0
