.PHONY: build dev mcp tidy clean up down logs rebuild smoke-test ps

# --- Local Development ---
build:
	go build -o bin/server ./cmd/server
	go build -o bin/mcp ./cmd/mcp

dev:
	go run ./cmd/server

mcp:
	go run ./cmd/mcp

tidy:
	go mod tidy

clean:
	rm -rf bin/

# --- Docker Compose ---
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

rebuild:
	docker compose build --no-cache
	docker compose up -d

smoke-test:
	@echo "=== Health Check ==="
	@curl -sf http://localhost:8080/health && echo " ✓ Backend OK" || echo " ✗ Backend FAILED"
	@echo "=== Frontend ==="
	@curl -sf -o /dev/null http://localhost:3001/ && echo " ✓ Frontend OK" || echo " ✗ Frontend FAILED"
	@echo "=== Create Run ==="
	@curl -sf -X POST http://localhost:8080/api/v1/runs \
		-H "Content-Type: application/json" \
		-d '{"project_path":"/tmp/test","requirements":"smoke test"}' && echo " ✓ Create Run OK" || echo " ✗ Create Run FAILED"
	@echo "=== List Runs ==="
	@curl -sf http://localhost:8080/api/v1/runs && echo " ✓ List Runs OK" || echo " ✗ List Runs FAILED"
	@echo "=== Smoke Test Complete ==="
