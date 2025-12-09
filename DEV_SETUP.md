# TrailMemo API - Development Setup Guide

Complete guide for setting up and running the API locally with hot reload.

## Prerequisites

- Docker Desktop installed and running
- Go 1.21+ installed
- Air (for hot reload): `go install github.com/cosmtrek/air@latest`

## Quick Start

### 1. One-Command Setup

```bash
make dev-setup
```

This will:
- ✅ Create `.env` from template
- ✅ Start PostgreSQL in Docker
- ✅ Run database migrations
- ✅ Set up everything for development

### 2. Edit Your Environment

```bash
# Edit .env with your Firebase credentials
nano .env
```

Required values:
- `FIREBASE_PROJECT_ID` - Your Firebase project ID
- `FIREBASE_STORAGE_BUCKET` - Your storage bucket name
- `FIREBASE_SERVICE_ACCOUNT_PATH` - Path to your service account JSON

### 3. Start Development Server

```bash
make dev
```

The API starts at `http://localhost:8080` with **hot reload** - any code changes will automatically rebuild! 🔥

## Development Commands

### Database Management

```bash
# Start Docker containers (PostgreSQL)
make dev-start

# Stop containers
make dev-stop

# View container logs
make dev-logs

# Connect to database directly
make dev-db

# Reset database (WARNING: deletes all data)
make dev-reset

# Start pgAdmin (database GUI)
make dev-pgadmin
# Then open: http://localhost:5050
# Email: admin@trailmemo.local
# Password: admin
```

### Application Commands

```bash
# Run with hot reload
make dev

# Run without hot reload
make run

# Build binary
make build

# Run tests
make test

# Run migrations
make migrate
```

## Development Workflow

### Daily Development

```bash
# 1. Start Docker containers
make dev-start

# 2. Start development server with hot reload
make dev

# 3. Make changes to code - it rebuilds automatically!

# 4. When done, stop containers
make dev-stop
```

### Hot Reload in Action

When you save any `.go` file:
```
🔥 File changed: internal/handlers/memos.go
🔨 Rebuilding...
✅ Build successful!
🚀 Restarting server...
```

### Database Access

**CLI Access:**
```bash
make dev-db
# Password: trailmemo_dev_password
```

**pgAdmin (GUI):**
```bash
make dev-pgadmin
# Open: http://localhost:5050
# Login: admin@trailmemo.local / admin

# Add server:
# Host: postgres (or host.docker.internal on Mac)
# Port: 5432
# Database: trailmemo_dev
# User: trailmemo
# Password: trailmemo_dev_password
```

## Environment Configuration

### Local Development (.env)

```bash
PORT=8080
ENV=development
DATABASE_URL=postgresql://trailmemo:trailmemo_dev_password@localhost:5432/trailmemo_dev?sslmode=disable
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_STORAGE_BUCKET=your-project-id.appspot.com
FIREBASE_SERVICE_ACCOUNT_PATH=./serviceAccountKey.json
JWT_SECRET=dev_secret_here
MAX_UPLOAD_SIZE=52428800
```

### Production (Railway)

Production uses Railway environment variables - never commit production credentials!

## Testing Locally

### 1. Get Firebase Token

```bash
WEB_API_KEY="your_firebase_web_api_key"

# Create test user
TOKEN=$(curl -s -X POST "https://identitytoolkit.googleapis.com/v1/accounts:signUp?key=$WEB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@test.com","password":"Test123!","returnSecureToken":true}' \
  | grep -o '"idToken":"[^"]*' | cut -d'"' -f4)

echo "Token: $TOKEN"
```

### 2. Test Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Dev User","department":"Development"}'

# Create memo
curl -X POST http://localhost:8080/api/v1/memos \
  -H "Authorization: Bearer $TOKEN" \
  -F "text=Testing local development" \
  -F "duration_seconds=10" \
  -F "latitude=45.6789" \
  -F "longitude=-111.0123" \
  -F "park_name=Dev Park"

# List memos
curl http://localhost:8080/api/v1/memos \
  -H "Authorization: Bearer $TOKEN"
```

## Troubleshooting

### Docker containers won't start

```bash
# Check if Docker is running
docker ps

# Check container status
docker-compose ps

# View logs
make dev-logs
```

### Port 5432 already in use

```bash
# Stop existing PostgreSQL
brew services stop postgresql

# Or use different port in docker-compose.yml
# Change "5432:5432" to "5433:5432"
# Update DATABASE_URL to use port 5433
```

### Hot reload not working

```bash
# Make sure Air is installed
air -v

# If not installed:
go install github.com/cosmtrek/air@latest

# Clear tmp directory
rm -rf tmp/
```

### Database connection errors

```bash
# Check PostgreSQL is running
docker-compose ps

# Try connecting directly
make dev-db

# Reset database if needed
make dev-reset
```

### "No .env file found"

```bash
# Copy the example
cp env.development.example .env

# Edit with your values
nano .env
```

## Project Structure

```
trailmemo-api/
├── cmd/server/main.go          # Entry point
├── internal/                    # Application code
│   ├── handlers/               # HTTP handlers
│   ├── models/                 # Data models
│   ├── repository/             # Database layer
│   ├── services/               # External services
│   └── middleware/             # HTTP middleware
├── config/                      # Configuration
├── migrations/                  # Database migrations
├── docker-compose.yml           # Local development services
├── .env                        # Local environment (git ignored)
├── env.development.example     # Environment template
└── .air.toml                   # Hot reload config
```

## Tips & Tricks

### Keep Production and Development Separate

```bash
# Development .env (local PostgreSQL)
DATABASE_URL=postgresql://trailmemo:trailmemo_dev_password@localhost:5432/trailmemo_dev

# Production stays in Railway (never in .env)
```

### Watch Log Output

```bash
# Terminal 1: Docker logs
make dev-logs

# Terminal 2: API logs
make dev
```

### Quick Database Reset

```bash
# If your database gets messy during development
make dev-reset
```

### Use pgAdmin for Complex Queries

```bash
make dev-pgadmin
# Visual query builder and data browser at http://localhost:5050
```

## Next Steps

- ✅ Development environment running
- ✅ Hot reload working
- ✅ Database accessible
- 👉 Start building features!
- 👉 Test with curl or Postman
- 👉 Build iOS app integration

## Common Development Tasks

### Add a New Migration

```sql
-- Create: migrations/002_add_feature.sql
ALTER TABLE memos ADD COLUMN some_field TEXT;
```

```bash
# Run it
make migrate
```

### Test a Handler Change

1. Edit `internal/handlers/memos.go`
2. Save - Air automatically rebuilds
3. Test with curl
4. See changes immediately!

### Debug Database

```bash
# Connect to DB
make dev-db

# List tables
\dt

# Query data
SELECT * FROM memos LIMIT 5;

# Exit
\q
```

---

**Happy Developing! 🚀**

For production deployment, see [DEPLOYMENT.md](documentation/DEPLOYMENT.md)

