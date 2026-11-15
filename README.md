# ShopFlow

Event-driven order processing: an HTTP **API** (Go, Gin) stores orders in **PostgreSQL**
and publishes `ORDER_CREATED` events to **RabbitMQ**; a **worker** consumes them.

```
client -> API --(one SQL statement)--> orders, order_items, stock, outbox_events   (PostgreSQL)
                 |
                 +-- outbox relay (in the API, 1s poll) --> RabbitMQ "orders" queue --> worker
                                                                  \-> "orders.dlq" (failed messages)
```

## Quick start

Requires Docker with Compose v2.

```bash
make up            # build + start everything, waits until healthy
make test-order    # create an order
make smoke         # end-to-end checks (HTTP, DB state, outbox, worker)
make logs
make down          # stop AND delete data
```

API: `GET /healthz`, `GET /products`, `POST /orders` with `{"productId":1,"quantity":2}`

| Status | Meaning |
|--------|---------|
| 201 | order created |
| 400 | invalid JSON, `productId <= 0` or `quantity <= 0` |
| 404 | unknown product |
| 409 | not enough stock |

RabbitMQ UI: http://localhost:15672 (guest / guest)

## Design notes

- **One atomic SQL statement** reserves stock, creates the order + item and writes the
  outbox event. Row locking means concurrent orders can never oversell.
  Money math happens in Postgres `NUMERIC`, not floats.
- **Transactional outbox**: an order can never exist without its event. Delivery to
  RabbitMQ is **at-least-once** (publisher confirms; unpublished events are retried).
- **Worker** acks only after success, retries a failing message once, then moves it to
  `orders.dlq`. Handlers must be idempotent.
- Migrations in `migrations/` run on a fresh Postgres volume only. For an existing
  database run `002_constraints_and_outbox.sql` manually (it is idempotent) or `make down`.

## Tests

```bash
make test               # unit tests, no services needed
make up && make test-integration
```

Integration tests skip themselves unless `TEST_DATABASE_URL` / `TEST_RABBITMQ_URL` are set.

## Roadmap

Kubernetes/Helm, ArgoCD, Crossplane (AWS/Azure/GCP) — scaffolding lives in `deploy/`,
`platform/` and `infrastructure/` and is not functional yet.
