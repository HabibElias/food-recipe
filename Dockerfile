# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Leverage docker layer caching for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary without debug symbols for small image size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /food-recipe ./cmd/app

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies (certificates for HTTPS/Cloudinary/Chapa)
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /food-recipe /app/food-recipe

# Non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8000

ENTRYPOINT ["/app/food-recipe"]
