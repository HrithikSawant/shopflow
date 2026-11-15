FROM golang:1.27 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shopflow-worker ./cmd/worker

FROM alpine:3.22

RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /app/shopflow-worker .
USER app

CMD ["./shopflow-worker"]
