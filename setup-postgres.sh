#!/bin/bash

echo "🚀 GoCommerce PostgreSQL Setup"
echo "=============================="
echo ""

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is not installed"
    echo "   Please install Docker and Docker Compose first"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "❌ Docker is not running"
    echo "   Please start Docker first"
    exit 1
fi

echo "1️⃣  Starting PostgreSQL container..."
docker-compose up -d

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 5

# Check if container is running
if ! docker-compose ps | grep -q "Up"; then
    echo "❌ Failed to start PostgreSQL"
    echo "   Check logs with: docker-compose logs postgres"
    exit 1
fi

echo "✓ PostgreSQL is running"
echo ""

echo "2️⃣  Running migrations..."
cd migrations/examples/postgresql

# Install dependencies if needed
if [ ! -d "vendor" ] && [ ! -f "go.sum" ]; then
    echo "   Installing dependencies..."
    go mod download
fi

# Run migrations
go run main.go

cd ../../..

echo ""
echo "✅ Setup Complete!"
echo ""
echo "📋 Connection Details:"
echo "   Host:     localhost"
echo "   Port:     5432"
echo "   Database: edomain"
echo "   Username: edomain"
echo "   Password: edomain"
echo ""
echo "💡 Useful Commands:"
echo "   View tables:     docker-compose exec postgres psql -U edomain -d edomain -c '\\dt'"
echo "   View migrations: docker-compose exec postgres psql -U edomain -d edomain -c 'SELECT * FROM gocommerce_migrations;'"
echo "   Connect to DB:   docker-compose exec postgres psql -U edomain -d edomain"
echo "   Stop database:   docker-compose down"
echo "   Remove data:     docker-compose down -v"
echo ""
