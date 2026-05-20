.PHONY: docker-up docker-down frontend-dev run-gateway run-user run-sales run-management fmt test tidy smoke

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

frontend-dev:
	cd frontend && npm install && npm run dev

run-gateway:
	cd api-gateway && go run ./cmd

run-user:
	cd user-service && go run ./cmd/server

run-sales:
	cd bakery-sales-service && go run ./cmd/server

run-management:
	cd bakery-management-service && go run ./cmd/server

fmt:
	gofmt -w shared api-gateway user-service bakery-sales-service bakery-management-service

tidy:
	cd shared && go mod tidy
	cd api-gateway && go mod tidy
	cd user-service && go mod tidy
	cd bakery-sales-service && go mod tidy
	cd bakery-management-service && go mod tidy

test:
	cd shared && go test ./...
	cd api-gateway && go test ./...
	cd user-service && go test ./...
	cd bakery-sales-service && go test ./...
	cd bakery-management-service && go test ./...

smoke:
	bash scripts/smoke-test.sh
