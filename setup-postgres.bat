@echo off
setlocal EnableDelayedExpansion

echo 🚀 GoCommerce PostgreSQL Setup
echo ==============================
echo.

REM Check if docker-compose is installed
docker-compose --version >nul 2>&1
if errorlevel 1 (
    echo ❌ docker-compose is not installed
    echo    Please install Docker and Docker Compose first
    exit /b 1
)

REM Check if Docker is running
docker info >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker is not running
    echo    Please start Docker first
    exit /b 1
)

echo 1️⃣  Starting PostgreSQL container...
docker-compose up -d

REM Wait for PostgreSQL to be ready
echo ⏳ Waiting for PostgreSQL to be ready...
timeout /t 5 /nobreak >nul

REM Check if container is running
docker-compose ps | findstr "Up" >nul
if errorlevel 1 (
    echo ❌ Failed to start PostgreSQL
    echo    Check logs with: docker-compose logs postgres
    exit /b 1
)

echo ✓ PostgreSQL is running
echo.

echo 2️⃣  Running migrations...
cd migrations\examples\postgresql

REM Check if dependencies need to be downloaded
if not exist "go.sum" (
    echo    Installing dependencies...
    go mod download
)

REM Run migrations
go run main.go

cd ..\..\..

echo.
echo ✅ Setup Complete!
echo.
echo 📋 Connection Details:
echo    Host:     localhost
echo    Port:     5432
echo    Database: edomain
echo    Username: edomain
echo    Password: edomain
echo.
echo 💡 Useful Commands:
echo    View tables:     docker-compose exec postgres psql -U edomain -d edomain -c "\dt"
echo    View migrations: docker-compose exec postgres psql -U edomain -d edomain -c "SELECT * FROM gocommerce_migrations;"
echo    Connect to DB:   docker-compose exec postgres psql -U edomain -d edomain
echo    Stop database:   docker-compose down
echo    Remove data:     docker-compose down -v
echo.

endlocal
