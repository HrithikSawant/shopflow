up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

test-order:
	curl -X POST http://localhost:8080/orders \
	-H "Content-Type: application/json" \
	-d '{"productId":1,"quantity":2}'
