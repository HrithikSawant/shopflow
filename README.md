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
make up            # creates .env with random passwords on first run, builds, starts, waits until healthy
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

RabbitMQ UI: http://localhost:15672 (user `RABBITMQ_USER` and password `RABBITMQ_PASSWORD` from your `.env`)

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

## Configuration and secrets

- **No credentials are stored in the repository.** `make env` (run by `make up`) creates a
  git-ignored `.env` from `.env.example` with random passwords. `docker-compose.yaml`
  refuses to start if a value is missing.
- The API and worker **fail to start** without `DATABASE_URL` / `RABBITMQ_URL`; there are no
  built-in default passwords in the code.
- Postgres, RabbitMQ and the RabbitMQ UI are published on `127.0.0.1` only. The built-in
  RabbitMQ `guest` account is not created.
- CI generates fresh random passwords on every run and masks them in the logs. Every push is
  scanned for committed secrets with gitleaks.

### Before production

This setup is for local development and CI. For a real deployment:

- Keep secrets in a secret manager (AWS Secrets Manager / Vault) and deliver them with
  External Secrets Operator or Sealed Secrets, never as plain values in Helm values or Git.
- Use Docker/Kubernetes secrets or `POSTGRES_PASSWORD_FILE` instead of environment variables
  (env values are visible in `docker inspect`).
- Use TLS: `sslmode=verify-full` for Postgres and `amqps://` for RabbitMQ.
- Give the API its own least-privilege database role (no superuser / schema ownership);
  run migrations with a separate role.
- Rotate credentials and restrict network access with network policies / security groups.

## Tests

```bash
make test               # unit tests, no services needed
make up && make test-integration   # own database + RabbitMQ vhost: dev data is untouched
```

Integration tests skip themselves unless `TEST_DATABASE_URL` / `TEST_RABBITMQ_URL` are set.

