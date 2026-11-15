#!/usr/bin/env bash
# End-to-end smoke test. Run against a running stack (`make up`).
#
#   API      base URL of the API            (default http://localhost:8080)
#   PSQL     command that runs one SQL query and prints bare values
#            (default: psql inside the compose postgres container)
#   CHECK_WORKER=0  skip checking the worker's container logs
set -euo pipefail

# Pick up POSTGRES_USER / POSTGRES_DB from .env (no passwords are needed here:
# psql runs inside the container over its local socket).
if [ -f .env ]; then set -a; . ./.env; set +a; fi

API="${API:-http://localhost:8080}"
PSQL="${PSQL:-docker compose exec -T postgres psql -U ${POSTGRES_USER:-shopflow} -d ${POSTGRES_DB:-shopflow} -tA -c}"
CHECK_WORKER="${CHECK_WORKER:-1}"

BODY="$(mktemp)"
trap 'rm -f "$BODY"' EXIT

fail() { echo "FAIL: $*" >&2; exit 1; }

# expect_status <code> <curl args...>
expect_status() {
    local want="$1"; shift
    local got
    got="$(curl -s -o "$BODY" -w '%{http_code}' "$@" || true)"
    [ "$got" = "$want" ] || fail "expected HTTP $want, got $got: $(cat "$BODY")"
}

sql() { $PSQL "$1" | tr -d '[:space:]'; }

post_order() { # <json>
    curl -s -o "$BODY" -w '%{http_code}' -X POST "$API/orders" \
        -H "Content-Type: application/json" -d "$1" || true
}

curl -s -o /dev/null "$API/healthz" || fail "API is not reachable at $API - start the stack first with: make up"

echo "health + products"
expect_status 200 "$API/healthz"
expect_status 200 "$API/products"

echo "validation"
[ "$(post_order '{}')" = 400 ] || fail "empty body should be 400"
[ "$(post_order '{"productId":1,"quantity":0}')" = 400 ] || fail "quantity 0 should be 400"
[ "$(post_order '{"productId":1,"quantity":-5}')" = 400 ] || fail "negative quantity should be 400"
[ "$(post_order '{"productId":999999,"quantity":1}')" = 404 ] || fail "unknown product should be 404"
[ "$(post_order '{"productId":1,"quantity":1000000}')" = 409 ] || fail "oversell should be 409"

echo "create order"
stock_before="$(sql "select inventory_count from products where id = 1")"
[ "$(post_order '{"productId":1,"quantity":2}')" = 201 ] || fail "order should be 201: $(cat "$BODY")"
order_id="$(sed -n 's/.*"order":{"id":\([0-9]*\).*/\1/p' "$BODY")"
[ -n "$order_id" ] || fail "no order id in response: $(cat "$BODY")"

echo "database state for order $order_id"
ok="$(sql "select count(*) from orders o
           join order_items i on i.order_id = o.id
           join products p on p.id = i.product_id
           where o.id = $order_id and i.quantity = 2
             and i.price = p.price and o.total_amount = p.price * 2")"
[ "$ok" = 1 ] || fail "order total / unit price are wrong"

stock_after="$(sql "select inventory_count from products where id = 1")"
[ "$stock_after" = "$((stock_before - 2))" ] || fail "stock should drop by 2 ($stock_before -> $stock_after)"

echo "event delivery (outbox -> RabbitMQ)"
for _ in $(seq 1 20); do
    published="$(sql "select count(*) from outbox_events where order_id = $order_id and published_at is not null")"
    [ "$published" = 1 ] && break
    sleep 1
done
[ "${published:-0}" = 1 ] || fail "outbox event was not published within 20s"

if [ "$CHECK_WORKER" = 1 ]; then
    echo "worker processed the event"
    for _ in $(seq 1 20); do
        docker compose logs worker 2>&1 | grep -q "order $order_id: generating invoice" && found=1 && break
        sleep 1
    done
    [ "${found:-0}" = 1 ] || fail "worker did not process order $order_id"
fi

echo "Smoke tests passed"
