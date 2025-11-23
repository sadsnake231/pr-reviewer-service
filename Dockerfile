# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/service cmd/main.go

# Runtime stage
FROM golang:1.25-alpine

RUN apk --no-cache add ca-certificates
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PATH=$PATH:/root/go/bin

WORKDIR /root/

COPY --from=builder /app/service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./service"]
