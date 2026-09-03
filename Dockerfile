FROM mirror.gcr.io/library/golang:1.26.7-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    GOARCH=amd64 GIN_MODE=release \
    go build -ldflags "-s -w" \
    -o /loan-service ./cmd/server

FROM mirror.gcr.io/library/alpine:3.24 AS production

WORKDIR /app

# ENV USE_POSTGRES=true
# ENV POSTGRES_SERVER_HOST=172.19.0.2
# ENV POSTGRES_SERVER_PORT=5432
# ENV POSTGRES_DB=loan-engine
# ENV POSTGRES_USER=loan-engine
# ENV POSTGRES_PASSWORD=password
# ENV POSTGRES_TIMEOUT=10s

RUN apk add --no-cache ca-certificates

# Run as a non-root user
RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

COPY --from=builder /loan-service /app/loan-service

RUN chown appuser:appgroup /app/loan-service

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/loan-service"]