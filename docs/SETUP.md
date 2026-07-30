# GoTest Agent Setup Guide

Panduan lengkap untuk setup dan menjalankan GoTest Agent di environment development dan production.

## Prerequisites

### System Requirements
- **OS**: macOS, Linux, atau Windows (dengan WSL2)
- **CPU**: 4+ cores (recommended)
- **RAM**: 8GB minimum, 16GB recommended
- **Disk**: 20GB free space
- **Docker**: 24.0+ dengan Docker Compose v2
- **Git**: 2.30+

### Development Tools
- **Go**: 1.26.4+
- **Node.js**: 20+ (untuk frontend development)
- **npm**: 10+ (untuk frontend development)
- **PostgreSQL**: 16+ (optional, falls back to in-memory)
- **Redis**: 7+ (optional, falls back to in-memory queue)

## Quick Start (Docker Compose)

Cara tercepat untuk menjalankan GoTest Agent dengan semua services:

```bash
# Clone repository
git clone https://github.com/sanhaji182/agent_test.git
cd agent_test

# Copy environment template
cp .env.example .env

# Edit .env dan tambahkan API keys
# Minimal salah satu LLM provider harus dikonfigurasi
nano .env

# Start semua services
make up

# Verify semua services running
make smoke-test

# View logs
make logs
```

Services yang akan berjalan:
- **Backend**: http://localhost:8080 (API)
- **Frontend**: http://localhost:3001 (Dashboard)
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **Steel Browser**: http://localhost:3010
- **LangGraph Sidecar**: http://localhost:8000

## Environment Variables

### Required Variables

Minimal salah satu LLM provider harus dikonfigurasi:

```bash
# Anthropic (Claude)
ANTHROPIC_API_KEY=your-anthropic-api-key

# OpenAI
OPENAI_API_KEY=your-openai-api-key

# Google Gemini
GOOGLE_API_KEY=your-google-api-key

# DeepSeek
DEEPSEEK_API_KEY=your-deepseek-api-key
```

### Optional Variables

```bash
# LLM Configuration
LLM_MODEL=claude-sonnet-4-6  # Default model
LLM_PROVIDER=anthropic       # Default provider

# API Security
API_KEY=your-api-key         # API authentication (empty = no auth)
JWT_SECRET=your-jwt-secret   # JWT secret for authentication
GITHUB_WEBHOOK_SECRET=your-webhook-secret  # GitHub webhook verification

# Database
DATABASE_URL=postgres://user:password@localhost:5432/gotest_agent?sslmode=disable
POSTGRES_PASSWORD=your-strong-password  # Change in production!

# Redis
REDIS_URL=redis://localhost:6379

# Steel Browser
STEEL_API_URL=http://localhost:3010
STEEL_API_KEY=your-steel-api-key  # Optional

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3001  # Comma-separated origins

# LangGraph Sidecar
LANGGRAPH_SIDECAR_URL=http://localhost:8000
```

### Production Security Checklist

**WAJIB diubah untuk production:**

1. **Database credentials**:
   ```bash
   # Generate strong password
   openssl rand -base64 32
   ```

2. **API keys**:
   ```bash
   # Generate API key
   openssl rand -hex 32
   ```

3. **JWT secret**:
   ```bash
   # Generate JWT secret
   openssl rand -base64 64
   ```

4. **GitHub webhook secret**:
   ```bash
   # Generate webhook secret
   openssl rand -hex 32
   ```

5. **Bind ports to localhost**:
   ```yaml
   # docker-compose.yml
   postgres:
     ports:
       - "127.0.0.1:5432:5432"  # instead of "5432:5432"
   
   redis:
     ports:
       - "127.0.0.1:6379:6379"  # instead of "6379:6379"
   ```

## Local Development Setup

### Backend Development

```bash
# Install dependencies
go mod download

# Run backend
go run ./cmd/server

# Backend akan berjalan di http://localhost:8080
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# Frontend akan berjalan di http://localhost:3000
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run specific package tests
go test ./internal/parser/... -v
go test ./internal/ai/... -v
go test ./internal/agent/... -v
```

### Running Linter

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

## Docker Commands

### Basic Commands

```bash
# Start all services
make up

# Stop all services
make down

# View logs
make logs

# Rebuild and restart
make rebuild

# Show service status
make ps

# Run smoke test
make smoke-test
```

### Docker Compose Commands

```bash
# Start specific service
docker compose up backend

# Stop specific service
docker compose stop backend

# View logs for specific service
docker compose logs -f backend

# Rebuild specific service
docker compose up --build backend

# Remove all containers
docker compose down

# Remove all containers and volumes
docker compose down -v

# View resource usage
docker compose top
```

### Database Management

```bash
# Access PostgreSQL shell
docker compose exec postgres psql -U postgres -d gotest_agent

# Backup database
docker compose exec postgres pg_dump -U postgres gotest_agent > backup.sql

# Restore database
docker compose exec -T postgres psql -U postgres gotest_agent < backup.sql

# Reset database
docker compose exec postgres psql -U postgres -c "DROP DATABASE gotest_agent;"
docker compose exec postgres psql -U postgres -c "CREATE DATABASE gotest_agent;"
```

### Redis Management

```bash
# Access Redis CLI
docker compose exec redis redis-cli

# Flush Redis
docker compose exec redis redis-cli FLUSHALL

# Monitor Redis
docker compose exec redis redis-cli MONITOR
```

## Testing API Endpoints

### Using cURL

```bash
# Health check
curl http://localhost:8080/health

# Create test run
curl -X POST http://localhost:8080/api/v1/runs \
  -H "Content-Type: application/json" \
  -d '{
    "project_path": "/path/to/project",
    "requirements": "test login and checkout flows"
  }'

# List test runs
curl http://localhost:8080/api/v1/runs

# Get test run details
curl http://localhost:8080/api/v1/runs/{run_id}

# Get test run report
curl http://localhost:8080/api/v1/runs/{run_id}/report

# Cancel test run
curl -X POST http://localhost:8080/api/v1/runs/{run_id}/cancel

# Create schedule
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nightly-tests",
    "project_path": "/path/to/project",
    "frequency": "daily",
    "time": "02:00",
    "enabled": true
  }'

# Get metrics
curl http://localhost:8080/api/v1/metrics/summary
```

### Using HTTPie

```bash
# Install HTTPie
brew install httpie  # macOS
# or
pip install httpie  # Python

# Create test run
http POST http://localhost:8080/api/v1/runs \
  project_path=/path/to/project \
  requirements="test login flows"

# List test runs
http GET http://localhost:8080/api/v1/runs
```

### Using Postman

1. Import collection dari `docs/postman_collection.json` (jika ada)
2. Set base URL: `http://localhost:8080`
3. Set API key jika dikonfigurasi:
   - Key: `Authorization`
   - Value: `Bearer YOUR_API_KEY`

## Configuration Examples

### Multi-Provider Setup

```bash
# .env
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AIza...
DEEPSEEK_API_KEY=sk-...

# Default to Anthropic
LLM_PROVIDER=anthropic
LLM_MODEL=claude-sonnet-4-6

# Or use OpenAI
LLM_PROVIDER=openai
LLM_MODEL=gpt-5.5-pro

# Or use local LLM
LLM_PROVIDER=local
LLM_MODEL=llama-3.1-70b
LLM_BASE_URL=http://localhost:11434
```

### Production Database Setup

```bash
# .env
DATABASE_URL=postgres://gotest:STRONG_PASSWORD@postgres:5432/gotest_agent?sslmode=require
POSTGRES_PASSWORD=STRONG_RANDOM_PASSWORD

# Generate strong password
openssl rand -base64 32
```

### Rate Limiting Configuration

```bash
# .env
# Rate limiting is implemented in middleware
# Default: 100 requests per minute per IP
# To change, modify internal/api/ratelimit.go
```

### CORS Configuration

```bash
# .env
# Disable CORS (not recommended for production)
CORS_ALLOWED_ORIGINS=

# Allow specific origins
CORS_ALLOWED_ORIGINS=http://localhost:3001,https://your-domain.com

# Allow all origins (development only)
CORS_ALLOWED_ORIGINS=*
```

## Troubleshooting

### Backend Won't Start

**Problem**: Backend fails to start with database connection error

**Solution**:
```bash
# Check if PostgreSQL is running
docker compose ps postgres

# Restart PostgreSQL
docker compose restart postgres

# Check logs
docker compose logs postgres

# Reset database
docker compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS gotest_agent;"
docker compose exec postgres psql -U postgres -c "CREATE DATABASE gotest_agent;"
```

### Frontend Won't Start

**Problem**: Frontend fails to start with dependency errors

**Solution**:
```bash
cd frontend

# Clean install
rm -rf node_modules package-lock.json
npm install

# Clear Next.js cache
rm -rf .next

# Restart
npm run dev
```

### Port Already in Use

**Problem**: Port 8080 or 3001 already in use

**Solution**:
```bash
# Find process using port
lsof -i :8080  # macOS/Linux
# or
netstat -ano | findstr :8080  # Windows

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
# backend:
#   ports:
#     - "8081:8080"  # Change to 8081
```

### Playwright Browser Issues

**Problem**: Playwright fails to launch browser

**Solution**:
```bash
# Install Playwright browsers
npx playwright install

# Install specific browser
npx playwright install chromium

# Install with dependencies
npx playwright install --with-deps chromium

# Check installation
npx playwright install --dry-run
```

### Redis Connection Issues

**Problem**: Redis connection timeout

**Solution**:
```bash
# Check if Redis is running
docker compose ps redis

# Restart Redis
docker compose restart redis

# Check Redis logs
docker compose logs redis

# Test Redis connection
docker compose exec redis redis-cli ping
```

### LLM API Errors

**Problem**: LLM API returns errors

**Solution**:
```bash
# Check API key is set
echo $ANTHROPIC_API_KEY

# Test API key
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: YOUR_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-sonnet-4-6",
    "max_tokens": 100,
    "messages": [{"role": "user", "content": "Hello"}]
  }'

# Check rate limits
# Anthropic: https://console.anthropic.com/usage
# OpenAI: https://platform.openai.com/usage
```

### Test Execution Failures

**Problem**: Tests fail to execute

**Solution**:
```bash
# Check Steel Browser is running
docker compose ps steel-browser

# Restart Steel Browser
docker compose restart steel-browser

# Check Steel Browser logs
docker compose logs steel-browser

# Test Steel Browser API
curl http://localhost:3010/v1/sessions
```

### Database Migration Issues

**Problem**: Database migrations fail

**Solution**:
```bash
# Run migrations manually
docker compose exec postgres psql -U postgres -d gotest_agent -f /docker-entrypoint-initdb.d/init.sql

# Check migration status
docker compose exec postgres psql -U postgres -d gotest_agent -c "SELECT * FROM schema_migrations;"

# Reset and re-run migrations
docker compose exec postgres psql -U postgres -c "DROP DATABASE gotest_agent;"
docker compose exec postgres psql -U postgres -c "CREATE DATABASE gotest_agent;"
docker compose restart backend
```

## Monitoring

### Prometheus Metrics

```bash
# View Prometheus metrics
curl http://localhost:8080/metrics

# Key metrics:
# - gotest_runs_total: Total test runs
# - gotest_tests_total: Total tests executed
# - gotest_tests_passed: Tests passed
# - gotest_tests_failed: Tests failed
# - gotest_tests_healed: Tests healed
# - gotest_execution_duration_ms: Execution duration
```

### Jaeger Tracing

```bash
# Access Jaeger UI
open http://localhost:16686  # macOS
# or
xdg-open http://localhost:16686  # Linux

# Search traces for specific run
# Service: gotest-agent
# Operation: execute_run
```

### Logs

```bash
# View all logs
docker compose logs

# View specific service logs
docker compose logs backend
docker compose logs frontend
docker compose logs postgres

# Follow logs
docker compose logs -f backend

# View last 100 lines
docker compose logs --tail=100 backend
```

## Backup and Recovery

### Backup Database

```bash
# Create backup
docker compose exec postgres pg_dump -U postgres gotest_agent > backup_$(date +%Y%m%d).sql

# Compress backup
gzip backup_*.sql

# Restore from backup
docker compose exec -T postgres psql -U postgres gotest_agent < backup_20260115.sql
```

### Backup Configuration

```bash
# Backup .env
cp .env .env.backup.$(date +%Y%m%d)

# Backup docker-compose.yml
cp docker-compose.yml docker-compose.yml.backup.$(date +%Y%m%d)
```

### Recovery Procedure

```bash
# Stop all services
make down

# Restore database
docker compose exec -T postgres psql -U postgres gotest_agent < backup.sql

# Restore configuration
cp .env.backup.20260115 .env
cp docker-compose.yml.backup.20260115 docker-compose.yml

# Restart services
make up
```

## Production Deployment

### Using Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml gotest-agent

# View services
docker stack services gotest-agent

# Scale services
docker service scale gotest-agent_backend=3
```

### Using Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/

# View pods
kubectl get pods

# View logs
kubectl logs -f deployment/gotest-agent-backend

# Scale deployment
kubectl scale deployment/gotest-agent-backend --replicas=3
```

### Using Cloud Platforms

#### AWS ECS

```bash
# Create cluster
aws ecs create-cluster --cluster-name gotest-agent

# Register task definition
aws ecs register-task-definition --cli-input-json file://task-definition.json

# Create service
aws ecs create-service --cluster gotest-agent --service-name gotest-agent-backend --task-definition gotest-agent-backend --desired-count 3
```

#### Google Cloud Run

```bash
# Deploy service
gcloud run deploy gotest-agent-backend \
  --image gcr.io/PROJECT_ID/gotest-agent-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## Security Hardening

### TLS/SSL Setup

```bash
# Using Caddy (automatic HTTPS)
# Caddyfile
:443 {
  reverse_proxy backend:8080
  tls your-email@example.com
}

# Using nginx with Let's Encrypt
# /etc/nginx/sites-available/gotest-agent
server {
  listen 443 ssl;
  server_name your-domain.com;
  
  ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
  ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
  
  location / {
    proxy_pass http://backend:8080;
  }
}
```

### Firewall Configuration

```bash
# Allow only necessary ports
# UFW (Ubuntu)
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 22/tcp
ufw enable

# iptables
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -j DROP
```

### Monitoring and Alerting

```bash
# Set up Prometheus + Grafana
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'gotest-agent'
    static_configs:
      - targets: ['backend:8080']

# Grafana dashboard
# Import dashboard ID: 1860 (Node.js Application Dashboard)
```

## Best Practices

### Development

1. **Use environment variables** for all configuration
2. **Never commit secrets** to version control
3. **Use Docker Compose** for consistent development environment
4. **Run tests locally** before committing
5. **Use pre-commit hooks** for linting and formatting

### Production

1. **Change all default credentials** before deployment
2. **Use strong random secrets** for all secrets
3. **Bind database/Redis to localhost** only
4. **Use TLS/SSL** for all external connections
5. **Implement rate limiting** for API protection
6. **Use Docker secrets** for sensitive values
7. **Set up monitoring and alerting**
8. **Regular backups** of database and configuration
9. **Implement log aggregation** (ELK stack, Loki)
10. **Set up automated security scanning**

## Support

### Documentation
- [API Documentation](./API.md)
- [Architecture](./ARCHITECTURE.md)
- [Phase Plans](./PHASE-1-PLAN.md)

### Community
- GitHub Issues: https://github.com/sanhaji182/agent_test/issues
- GitHub Discussions: https://github.com/sanhaji182/agent_test/discussions

### Contact
- Email: support@gotest.ai
- Slack: #gotest-agent-support
