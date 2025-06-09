# Build stage
FROM golang:1.22.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o consumer .

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/consumer .
EXPOSE 8081
CMD ["./consumer"]
