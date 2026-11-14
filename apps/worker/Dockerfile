FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o shopflow-worker \
    ./cmd/worker

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/shopflow-worker .

CMD ["./shopflow-worker"]