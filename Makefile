env:
	@bash ./scripts/gen-env.sh

up: env
	docker compose up --build -d --wait

down: env
	docker compose down -v

logs: env
	docker compose logs -f

test:
	cd apps/api && go vet ./... && go test ./...
	cd apps/worker && go vet ./... && go test ./...

# needs `make up`. Runs against a separate <POSTGRES_DB>_test database (recreated
# with all migrations on every run) and a separate RabbitMQ vhost "<POSTGRES_DB>_test",
# so your development data and the live worker's queues are never touched.
# Passwords are read from .env inside the shell; the leading @ keeps them out of the log.
test-integration: env
	@set -a && . ./.env && set +a && \
	docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -q -U "$$POSTGRES_USER" -d postgres \
		-c "DROP DATABASE IF EXISTS $${POSTGRES_DB}_test" \
		-c "CREATE DATABASE $${POSTGRES_DB}_test" && \
	cat migrations/*.sql | docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -q \
		-U "$$POSTGRES_USER" -d "$${POSTGRES_DB}_test" >/dev/null && \
	(docker compose exec -T rabbitmq rabbitmqctl delete_vhost "$${POSTGRES_DB}_test" >/dev/null 2>&1 || true) && \
	docker compose exec -T rabbitmq rabbitmqctl add_vhost "$${POSTGRES_DB}_test" >/dev/null && \
	docker compose exec -T rabbitmq rabbitmqctl set_permissions -p "$${POSTGRES_DB}_test" "$$RABBITMQ_USER" ".*" ".*" ".*" >/dev/null && \
	(cd apps/api && \
		TEST_DATABASE_URL="postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$${POSTGRES_DB}_test" \
		TEST_RABBITMQ_URL="amqp://$$RABBITMQ_USER:$$RABBITMQ_PASSWORD@localhost:5672/$${POSTGRES_DB}_test" \
		go test ./... -p 1 -count=1) && \
	(cd apps/worker && \
		TEST_RABBITMQ_URL="amqp://$$RABBITMQ_USER:$$RABBITMQ_PASSWORD@localhost:5672/$${POSTGRES_DB}_test" \
		go test ./... -count=1)

smoke:
	bash ./tests/smoke.sh

test-order:
	curl -X POST http://localhost:8080/orders \
		-H "Content-Type: application/json" \
		-d '{"productId":1,"quantity":2}'
