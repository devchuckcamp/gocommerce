# Docker Compose Setup

This directory contains Docker configuration for running PostgreSQL with the gocommerce project.

## Quick Start

### Start PostgreSQL

```bash
docker-compose up -d
```

This will:
- Create a PostgreSQL 16 container
- Expose port 5432 (default PostgreSQL port)
- Create database `edomain`
- Set up user `edomain` with password `edomain`
- Persist data in a Docker volume

### Stop PostgreSQL

```bash
docker-compose down
```

### Stop and Remove Data

```bash
docker-compose down -v
```

## Connection Details

| Parameter | Value |
|-----------|-------|
| Host | `localhost` |
| Port | `5432` |
| Database | `edomain` |
| Username | `edomain` |
| Password | `edomain` |

## Connection Strings

### DSN (Data Source Name)
```
host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable
```

### Go Connection String
```go
db, err := sql.Open("postgres", 
    "host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable")
```

### Connection URL
```
postgresql://edomain:edomain@localhost:5432/edomain?sslmode=disable
```

## Docker Commands

### View Logs
```bash
docker-compose logs -f postgres
```

### Check Status
```bash
docker-compose ps
```

### Execute SQL
```bash
docker-compose exec postgres psql -U edomain -d edomain
```

### Backup Database
```bash
docker-compose exec postgres pg_dump -U edomain edomain > backup.sql
```

### Restore Database
```bash
docker-compose exec -T postgres psql -U edomain edomain < backup.sql
```

## Using with GoCommerce

### 1. Start PostgreSQL
```bash
docker-compose up -d
```

### 2. Run Migrations

Use the PostgreSQL migration example:

```bash
cd migrations/examples/postgresql
go run main.go
```

This will create 6 tables:
- `brands` - Product brands (Apple, Dell, etc.)
- `categories` - Product categories with hierarchy
- `products` - Products with brand/category relationships
- `carts` - Shopping carts
- `cart_items` - Cart line items
- `orders` - Customer orders

### 3. Seed Database

Populate with test data:

```bash
go run seed-products.go
```

This will create:
- 8 brands
- 8 categories
- 72 products (22 curated + 50 random)

### 4. View Data

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U edomain -d edomain

# View tables
\dt

# View products with brands
SELECT 
    p.name,
    p.base_price_amount / 100.0 as price,
    b.name as brand,
    c.name as category
FROM products p
JOIN brands b ON p.brand_id = b.id
JOIN categories c ON p.category_id = c.id
WHERE p.status = 'active'
LIMIT 10;
```

### 5. Run Sample Project

Update `sample-project/main.go` to use PostgreSQL instead of in-memory storage.

## Troubleshooting

### Port Already in Use

If port 5432 is already in use, change it in `docker-compose.yml`:

```yaml
ports:
  - "5432:5432"  # Use 5432 on host, 5432 in container
```

Then update connection strings to use port 5432.

### Container Won't Start

Check logs:
```bash
docker-compose logs postgres
```

### Permission Issues

If you get permission errors on Linux:
```bash
sudo chown -R $USER:$USER .
```

### Reset Database

Stop and remove volumes:
```bash
docker-compose down -v
docker-compose up -d
```

## Health Check

The PostgreSQL container includes a health check that runs every 10 seconds. Check status:

```bash
docker-compose ps
```

Healthy status should show:
```
NAME                  STATUS
gocommerce-postgres   Up X minutes (healthy)
```

## Environment Variables

You can override settings using a `.env` file:

```bash
# .env
POSTGRES_USER=myuser
POSTGRES_PASSWORD=mypassword
POSTGRES_DB=mydb
POSTGRES_PORT=5432
```

Then update `docker-compose.yml` to use variables:

```yaml
environment:
  POSTGRES_USER: ${POSTGRES_USER}
  POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
  POSTGRES_DB: ${POSTGRES_DB}
ports:
  - "${POSTGRES_PORT}:5432"
```

## Production Notes

⚠️ **This configuration is for development only!**

For production:
- Use strong passwords
- Enable SSL/TLS
- Use secrets management
- Configure proper networking
- Set up backups
- Use connection pooling
- Configure resource limits

## Next Steps

1. Start the database: `docker-compose up -d`
2. Run migrations (see examples in `migrations/examples/`)
3. Update sample-project to use PostgreSQL
4. Start building your e-commerce application!
