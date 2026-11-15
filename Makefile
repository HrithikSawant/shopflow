up:
	docker compose up --build -d --wait

down:
	docker compose down -v

logs:
	docker compose logs -f

test:
	cd apps/api && go vet ./... && go test ./...
	cd apps/worker && go vet ./... && go test ./...

# needs `make up` (Postgres + RabbitMQ on localhost)
test-integration:
	cd apps/api && TEST_DATABASE_URL=postgres://shopflow:shopflow@localhost:5432/shopflow \
		TEST_RABBITMQ_URL=amqp://guest:guest@localhost:5672/ go test ./... -p 1 -count=1
	cd apps/worker && TEST_RABBITMQ_URL=amqp://guest:guest@localhost:5672/ go test ./... -count=1

smoke:
	./tests/smoke.sh

test-order:
	curl -X POST http://localhost:8080/orders \
		-H "Content-Type: application/json" \
		-d '{"productId":1,"quantity":2}'
