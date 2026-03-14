.PHONY: dev dev-build test seed migrate-down lint clean

# Start all services with hot reload (backend via air, frontend via vite dev server)
dev:
	docker compose up --build

# Build without cache
dev-build:
	docker compose build --no-cache && docker compose up

# Run Go tests
test:
	cd backend && go test ./... -v

# Lint Go code
lint:
	cd backend && go vet ./...

# Seed sample mission data
seed:
	curl -s -X POST http://localhost:8080/api/missions \
		-H "Content-Type: application/json" \
		-d '{"name":"Summer Holiday 2026","icon":"sun","category":"travel","budget":3500,"currency":"EUR","locale":"fr-FR","constraints":[{"key":"max-flight","label":"Max flight duration","value":"4h","type":"hard"},{"key":"direct","label":"Direct from Paris","value":true,"type":"hard"},{"key":"kid-friendly","label":"Kid-friendly (ages 4-8)","value":true,"type":"hard"},{"key":"beach","label":"Beach access","value":true,"type":"soft"}],"costCategories":["Flights","Accommodation","Activities","Food","Transport"]}' \
		| jq .

# Roll back last migration (for development)
migrate-down:
	cd backend && go run cmd/migrate/main.go down 1

# Stop and remove containers + volumes
clean:
	docker compose down -v
