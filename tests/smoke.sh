#!/usr/bin/env bash

set -e

echo "Testing health..."

curl -fs http://localhost:8080/healthz

echo ""
echo "Testing products..."

curl -fs http://localhost:8080/products

echo ""
echo "Testing order creation..."

curl -fs \
-X POST http://localhost:8080/orders \
-H "Content-Type: application/json" \
-d '{"productId":1,"quantity":2}'

echo ""
echo "Smoke tests passed"
