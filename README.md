# BakePlan — Bakery Production and Return Management System

BakePlan is a Go microservices project for bakery production planning, client/store ordering, ingredient management, task control, statistics, and monitoring.

## Team distribution

| Member | Service |
|---|---|
| Bakyt | User Service |
| Arlan | Bakery Sales Service |
| Ersultan | Bakery Management Service |

## Architecture

```txt
Frontend → API Gateway → gRPC Microservices → PostgreSQL databases
                         ↓
                        NATS

Monitoring: API Gateway /metrics → Prometheus → Grafana
```

The frontend behaves like a real app. It does not show API Gateway URLs, health buttons, Grafana links, or developer demo instructions.

## What is included

- Go API Gateway
- Go User Service
- Go Bakery Sales Service
- Go Bakery Management Service
- gRPC communication using a JSON codec
- PostgreSQL database per service
- NATS event publishing/listening
- Redis cache used by API Gateway for products, baking plans, available products, and daily statistics
- React + TypeScript frontend
- Role-based UI:
  - Admin: dashboard/statistics, products/plans, orders, ingredients, tasks
  - Client/store: one store page only
- Popup/toast error handling
- Quantity validation for orders
- Prometheus and Grafana infrastructure
- Smoke test script
- Go unit tests for Sales and Management use cases
- SMTP email sender with safe dry-run mode when SMTP is not configured
- Email logs endpoint for verification

## Required installations

Install these before running the project:

1. **Docker**
2. **Docker Compose plugin**
3. **Git**
4. **Node.js 20+** only if you want to run the frontend locally outside Docker
5. **Go 1.23+** only if you want to run services locally outside Docker
6. **curl** and **python3** for the smoke test script

Check versions:

```bash
docker --version
docker compose version
node -v
npm -v
go version
python3 --version
curl --version
```

## Run with Docker

From the project root:

```bash
docker compose down -v
docker compose up --build
```

Use `down -v` for the first run of this updated version because the database schema now includes ingredient recipe sections.

Open the app:

```txt
http://localhost:3000
```

Other service URLs for developers only:

```txt
API Gateway: http://localhost:8080
Prometheus: http://localhost:9090
Grafana: http://localhost:3001
```

Grafana login:

```txt
admin / admin
```

## First login

Create an admin account from the Register tab:

```txt
Name: Admin
Email: admin@bakeplan.kz
Password: 123456
Role: Admin
```

Create a client/store account from the Register tab:

```txt
Name: Store A
Email: storea@bakeplan.kz
Password: 123456
Role: Client / Store
```

## Test that everything runs smoothly

Keep Docker running in one terminal:

```bash
docker compose up --build
```

Open a second terminal and run:

```bash
bash scripts/smoke-test.sh
```

Expected ending:

```txt
✅ Smoke test passed. Backend, auth, products, cache, plans, orders, order status, validation, ingredient-to-product sync, email logs, tasks, roles, and metrics are working.
```

You can also check containers:

```bash
docker compose ps
```

Check backend health:

```bash
curl http://localhost:8080/health
```

Expected:

```json
{"service":"api-gateway","status":"ok"}
```

Check metrics:

```bash
curl http://localhost:8080/metrics | head
```

Check logs if something fails:

```bash
docker compose logs api-gateway --tail=100
docker compose logs user-service --tail=100
docker compose logs bakery-sales-service --tail=100
docker compose logs bakery-management-service --tail=100
```

## Frontend pages

### Public

- Login
- Register

### Admin

- Dashboard + Statistics combined
- Products & Baking Plans combined
- Orders
- Ingredients / recipe base
- Tasks

### Client/store

- One page only:
  - available products
  - create order
  - my orders

## Important behavior

- Baking plans are created only by selecting an existing product.
- If order quantity is higher than available quantity, backend returns a conflict error and frontend shows a popup.
- Client users cannot access admin backend routes.
- Monitoring still exists in Docker Compose, Prometheus, and Grafana, but it is not shown in the user app.
- Redis cache is transparent for users. You can verify it with the `X-Cache` response header on cached GET endpoints.
- Creating an ingredient/recipe base also creates a product with the same name and cost, so it appears in Products & Baking Plans.
- Admin can move orders from `PENDING` to `CONFIRMED`, `DELIVERED`, or `CANCELLED`.


## Email setup

Email sending is implemented in the Management Service. By default, `SMTP_HOST` is empty, so the app saves the email to `email_logs` with `DRY_RUN` status. This keeps Docker safe for local testing.

To send real email through Gmail SMTP, create a `.env` file or export these variables before `docker compose up --build`:

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email@gmail.com
SMTP_PASSWORD=your_google_app_password
SMTP_FROM=your_email@gmail.com
```

Then test:

```bash
curl -X POST http://localhost:8080/admin/email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{"to":"client@example.com","subject":"Order confirmed","body":"Your bakery order was confirmed."}'
```

Check logs:

```bash
curl http://localhost:8080/admin/email-logs -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## Cache check

Run the same cached GET request twice. The second response should show `X-Cache: HIT`:

```bash
curl -i http://localhost:8080/products
curl -i http://localhost:8080/products
```

Cached endpoints:

```txt
GET /products
GET /bake-plans?date=YYYY-MM-DD
GET /available-products?date=YYYY-MM-DD
GET /admin/statistics/daily?date=YYYY-MM-DD
```

## Useful commands

Rebuild from zero:

```bash
docker compose down -v
docker compose build --no-cache
docker compose up
```

Build only frontend locally:

```bash
cd frontend
npm install
npm run build
```

Run frontend locally:

```bash
cd frontend
npm install
npm run dev
```

Run Go tests locally after dependencies are downloaded:

```bash
go work sync
cd shared && go mod tidy && go test ./...
cd ../api-gateway && go mod tidy && go test ./...
cd ../user-service && go mod tidy && go test ./...
cd ../bakery-sales-service && go mod tidy && go test ./...
cd ../bakery-management-service && go mod tidy && go test ./...
```

