# Build stage
FROM golang:1.23.4-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o auth-api .

# Run stage
FROM gcr.io/distroless/static-debian11
WORKDIR /
COPY --from=builder /app/auth-api /auth-api
EXPOSE 8082
CMD ["/auth-api"]