FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shopflow-api ./cmd/server

FROM alpine:3.24

RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=builder /app/shopflow-api .
USER app

ENV GIN_MODE=release

EXPOSE 8080

CMD ["./shopflow-api"]
