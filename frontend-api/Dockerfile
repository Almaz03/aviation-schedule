# Build stage
FROM golang:1.23.4-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o frontend-api .

# Run stage
FROM gcr.io/distroless/static-debian11
WORKDIR /
COPY --from=builder /app/frontend-api /frontend-api
EXPOSE 8083
CMD ["/frontend-api"]