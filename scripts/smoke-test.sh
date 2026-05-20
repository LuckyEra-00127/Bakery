#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
TS="$(date +%s)"
ADMIN_EMAIL="admin_${TS}@bakeplan.kz"
CLIENT_EMAIL="store_${TS}@bakeplan.kz"
DATE="2030-10-20"

json_get() {
  python3 -c "import json,sys; data=json.load(sys.stdin); print($1)"
}

echo "Checking API health..."
curl -fsS "$BASE_URL/health" >/dev/null

echo "Registering admin..."
ADMIN_REGISTER=$(curl -fsS -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"full_name\":\"Smoke Admin\",\"email\":\"$ADMIN_EMAIL\",\"password\":\"123456\",\"role\":\"ADMIN\"}")
ADMIN_TOKEN=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])' <<<"$ADMIN_REGISTER")
AUTH_HEADER="Authorization: Bearer $ADMIN_TOKEN"

echo "Creating product..."
PRODUCT_RESPONSE=$(curl -fsS -X POST "$BASE_URL/admin/products" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d '{"name":"Bread","price":200}')
PRODUCT_ID=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["product"]["id"])' <<<"$PRODUCT_RESPONSE")

echo "Checking products cache MISS then HIT..."
CACHE_FIRST=$(curl -s -D - "$BASE_URL/products" -o /tmp/bakeplan_products_1.json | tr -d '\r' | awk '/^X-Cache:/ {print $2}')
CACHE_SECOND=$(curl -s -D - "$BASE_URL/products" -o /tmp/bakeplan_products_2.json | tr -d '\r' | awk '/^X-Cache:/ {print $2}')
if [[ "${CACHE_SECOND:-}" != "HIT" ]]; then
  echo "Expected X-Cache HIT on second /products call, got '${CACHE_SECOND:-empty}'"
  exit 1
fi

echo "Creating baking plan..."
PLAN_RESPONSE=$(curl -fsS -X POST "$BASE_URL/admin/bake-plans" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d "{\"product_id\":\"$PRODUCT_ID\",\"plan_date\":\"$DATE\",\"planned_quantity\":300}")
PLAN_ID=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["bake_plan"]["id"])' <<<"$PLAN_RESPONSE")

echo "Creating order..."
ORDER_RESPONSE=$(curl -fsS -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d "{\"store_name\":\"Smoke Store\",\"bake_plan_id\":\"$PLAN_ID\",\"quantity\":30}")
ORDER_ID=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["order"]["id"])' <<<"$ORDER_RESPONSE")

echo "Accepting order..."
STATUS_RESPONSE=$(curl -fsS -X PATCH "$BASE_URL/admin/orders/status" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d "{\"id\":\"$ORDER_ID\",\"status\":\"CONFIRMED\"}")
STATUS=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["order"]["status"])' <<<"$STATUS_RESPONSE")
if [[ "$STATUS" != "CONFIRMED" ]]; then
  echo "Expected CONFIRMED order status, got $STATUS"
  exit 1
fi

echo "Checking insufficient quantity error handling..."
HTTP_CODE=$(curl -s -o /tmp/bakeplan_too_many.json -w "%{http_code}" -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d "{\"store_name\":\"Smoke Store\",\"bake_plan_id\":\"$PLAN_ID\",\"quantity\":9999}")
if [[ "$HTTP_CODE" != "409" ]]; then
  echo "Expected 409 for insufficient quantity, got $HTTP_CODE"
  cat /tmp/bakeplan_too_many.json
  exit 1
fi

echo "Creating ingredient with sections and checking it appears as product..."
INGREDIENT_NAME="Croissant_${TS}"
curl -fsS -X POST "$BASE_URL/admin/ingredients" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d "{\"name\":\"$INGREDIENT_NAME\",\"unit\":\"1 pc\",\"cost_per_unit\":350,\"sections\":[{\"name\":\"Butter\",\"amount\":\"40\",\"unit\":\"gr\"},{\"name\":\"Flour\",\"amount\":\"120\",\"unit\":\"gr\"}]}" >/dev/null
curl -fsS "$BASE_URL/products" | grep -q "$INGREDIENT_NAME"

echo "Sending email through management service..."
EMAIL_RESPONSE=$(curl -fsS -X POST "$BASE_URL/admin/email" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d '{"to":"client@example.com","subject":"BakePlan smoke test","body":"Your order is confirmed."}')
EMAIL_MESSAGE=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["message"])' <<<"$EMAIL_RESPONSE")
if [[ -z "$EMAIL_MESSAGE" ]]; then
  echo "Expected email message, got empty response"
  exit 1
fi
curl -fsS "$BASE_URL/admin/email-logs" -H "$AUTH_HEADER" | grep -q "BakePlan smoke test"

echo "Creating and updating task..."
TASK_RESPONSE=$(curl -fsS -X POST "$BASE_URL/admin/tasks" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d '{"title":"Clean oven","due_date":"2030-10-20"}')
TASK_ID=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["task"]["id"])' <<<"$TASK_RESPONSE")
curl -fsS -X PATCH "$BASE_URL/admin/tasks/status" \
  -H "Content-Type: application/json" -H "$AUTH_HEADER" \
  -d "{\"id\":\"$TASK_ID\",\"title\":\"Clean oven\",\"status\":\"IN_PROGRESS\",\"due_date\":\"2030-10-20\"}" >/dev/null
curl -fsS -X DELETE "$BASE_URL/admin/tasks/$TASK_ID" -H "$AUTH_HEADER" >/dev/null

echo "Registering client and checking forbidden admin access..."
CLIENT_REGISTER=$(curl -fsS -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"full_name\":\"Smoke Store\",\"email\":\"$CLIENT_EMAIL\",\"password\":\"123456\",\"role\":\"CLIENT\"}")
CLIENT_TOKEN=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["token"])' <<<"$CLIENT_REGISTER")
CLIENT_CODE=$(curl -s -o /tmp/bakeplan_forbidden.json -w "%{http_code}" "$BASE_URL/admin/tasks" -H "Authorization: Bearer $CLIENT_TOKEN")
if [[ "$CLIENT_CODE" != "403" ]]; then
  echo "Expected 403 when client opens admin route, got $CLIENT_CODE"
  cat /tmp/bakeplan_forbidden.json
  exit 1
fi

echo "Checking metrics endpoint..."
curl -fsS "$BASE_URL/metrics" | grep -q "bakeplan_gateway_requests_total"

echo "✅ Smoke test passed. Backend, auth, products, cache, plans, orders, order status, validation, ingredient-to-product sync, email logs, tasks, roles, and metrics are working."
