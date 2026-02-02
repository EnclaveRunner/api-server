# Build API-Server executable
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o /app/api-server .

# Create a minimal runtime image
FROM alpine:3.23

RUN apk --no-cache add ca-certificates
WORKDIR /app

COPY --from=builder /app/api-server .

EXPOSE 8080

CMD ["./api-server"]