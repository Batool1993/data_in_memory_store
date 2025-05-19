# builder stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build the server; if your main.go lives in cmd/server, point there
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server

# runtime stage
FROM alpine:latest
COPY --from=builder /bin/server /bin/server

# Provide STORE_SERVER so config.Load() doesn’t error out,
# plus the other TTL/auth settings
ENV STORE_SERVER=http://localhost:8080 \
    STORE_DEFAULT_TTL=60s \
    CLEANUP_INTERVAL=300s \
    STORE_API_TOKEN=my-secret-token

EXPOSE 8080
ENTRYPOINT ["/bin/server"]
