# ─────────────────────────────────────────────────────────────────────────────
# Scouter — Makefile
# ─────────────────────────────────────────────────────────────────────────────
# Profiles available in docker-compose.yml:
#   (default)   postgres + backend + frontend
#   seed        one-shot seed container (runs scouter-seed then exits)
#   monitoring  prometheus + grafana + cadvisor
#
# Usage examples:
#   make up              # start core stack
#   make up-seed         # start core + seed sample data
#   make up-monitoring   # start core + Grafana/Prometheus
#   make up-full         # start everything
#   make test-all        # run backend + frontend + e2e tests
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: \
  up up-seed up-monitoring up-full down restart logs ps \
  build build-backend build-frontend build-no-cache \
  seed \
  certs \
  test test-frontend test-coverage test-e2e test-all \
  lint lint-frontend typecheck \
  migrate-down \
  clean clean-volumes \
  help

# ── Stack ─────────────────────────────────────────────────────────────────────

DC := docker compose -f deployment/docker-compose.yml

## Start core stack (postgres + backend + frontend)
up:
	$(DC) up -d

## Start core stack + seed sample data, then follow logs
up-seed:
	$(DC) up -d
	$(DC) --profile seed up seed
	$(DC) logs -f backend frontend

## Start core stack + monitoring (Prometheus / Grafana / cAdvisor)
up-monitoring:
	$(DC) --profile monitoring up -d

## Start everything: core + seed + monitoring
up-full:
	$(DC) --profile seed --profile monitoring up -d
	$(DC) --profile seed up seed

## Stop all running containers (all profiles)
down:
	$(DC) --profile seed --profile monitoring down

## Restart core services
restart:
	$(DC) restart backend frontend

## Follow logs for core services (Ctrl-C to stop)
logs:
	$(DC) logs -f postgres backend frontend

## Show status of all containers
ps:
	$(DC) --profile seed --profile monitoring ps

# ── Build ─────────────────────────────────────────────────────────────────────

## Build all Docker images (uses layer cache)
build:
	$(DC) build

## Build only the backend image
build-backend:
	$(DC) build backend

## Build only the frontend image
build-frontend:
	$(DC) build frontend

## Force rebuild all images without cache, then start
build-no-cache:
	$(DC) build --no-cache
	$(DC) up -d

# ── TLS Certificates ──────────────────────────────────────────────────────────

## Generate local CA + wildcard cert for *.dev.local (requires openssl)
## Output: deployment/certs/ca.crt  deployment/certs/dev.local.crt  deployment/certs/dev.local.key
## Install deployment/certs/ca.crt in Windows: certutil -addstore -f "ROOT" ca.crt (Admin PowerShell)
certs:
	mkdir -p deployment/certs
	@echo "Generating CA key and certificate..."
	openssl genrsa -out deployment/certs/ca.key 4096
	openssl req -new -x509 -days 3650 -key deployment/certs/ca.key -out deployment/certs/ca.crt \
	  -subj "/CN=Scouter Dev CA/O=Scouter Dev"
	@echo "Generating wildcard key and CSR for *.dev.local..."
	openssl genrsa -out deployment/certs/dev.local.key 2048
	openssl req -new -key deployment/certs/dev.local.key -out deployment/certs/dev.local.csr \
	  -subj "/CN=*.dev.local"
	@echo "subjectAltName=DNS:*.dev.local,DNS:dev.local" > deployment/certs/dev.local.ext
	@echo "basicConstraints=CA:FALSE" >> deployment/certs/dev.local.ext
	@echo "keyUsage=digitalSignature,keyEncipherment" >> deployment/certs/dev.local.ext
	@echo "Signing certificate with SAN..."
	openssl x509 -req -days 825 \
	  -in deployment/certs/dev.local.csr \
	  -CA deployment/certs/ca.crt -CAkey deployment/certs/ca.key -CAcreateserial \
	  -out deployment/certs/dev.local.crt \
	  -extfile deployment/certs/dev.local.ext
	rm -f deployment/certs/dev.local.csr deployment/certs/dev.local.ext deployment/certs/ca.srl
	@echo ""
	@echo "Done. Files created:"
	@echo "  deployment/certs/ca.crt        <- install this in Windows (certutil -addstore -f ROOT ca.crt)"
	@echo "  deployment/certs/dev.local.crt <- mounted into Traefik"
	@echo "  deployment/certs/dev.local.key <- mounted into Traefik"

# ── Seed ─────────────────────────────────────────────────────────────────────

## Run the seed container (inserts home-server-2026 + summer-holiday-2026 missions)
seed:
	$(DC) --profile seed up seed

# ── Tests ────────────────────────────────────────────────────────────────────

## Run Go backend tests
test:
	cd backend && go test ./... -v

## Run frontend unit tests (vitest)
test-frontend:
	cd frontend && npm run test

## Run frontend unit tests with coverage report
test-coverage:
	cd frontend && npm run test:coverage

## Run Playwright E2E tests (requires frontend dev server on :5173)
test-e2e:
	cd frontend && npm run test:e2e

## Run all tests: backend + frontend unit + e2e
test-all: test test-frontend test-e2e

# ── Lint / Typecheck ──────────────────────────────────────────────────────────

## Lint Go code with go vet
lint:
	cd backend && go vet ./...

## Lint frontend with eslint
lint-frontend:
	cd frontend && npm run lint

## TypeScript typecheck (no emit)
typecheck:
	cd frontend && npm run typecheck

# ── Database ──────────────────────────────────────────────────────────────────

## Roll back the last migration (development only)
migrate-down:
	cd backend && go run cmd/migrate/main.go down 1

# ── Cleanup ───────────────────────────────────────────────────────────────────

## Stop containers but keep volumes
clean:
	$(DC) --profile seed --profile monitoring down

## Stop containers AND remove all volumes (destroys database data)
clean-volumes:
	$(DC) --profile seed --profile monitoring down -v

# ── Help ──────────────────────────────────────────────────────────────────────

## Show this help message
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Stack"
	@grep -E '^## ' Makefile | grep -A1 'up\|down\|restart\|logs\|ps' | sed 's/## /  /' | head -20
	@echo ""
	@awk '/^## /{desc=$$0; next} /^[a-zA-Z_-]+:/{printf "  \033[36m%-20s\033[0m %s\n", $$1, substr(desc,4)}' Makefile
	@echo ""
