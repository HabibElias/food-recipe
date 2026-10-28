# Stage 1: Build stage
FROM golang:1.26.10-alpine AS builder

WORKDIR /app

# Use the checked-in dependencies and the toolchain required by go.mod.
# Building needs no package-registry access or automatic toolchain downloads.
ENV GOTOOLCHAIN=local
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build statically linked binary without debug symbols for small image size
RUN CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-w -s" -o /food-recipe ./cmd/app

# Stage 2: Minimal runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# The runtime includes CA certificates and timezone data, and runs as non-root.
COPY --from=builder /food-recipe /app/food-recipe

EXPOSE 8000

ENTRYPOINT ["/app/food-recipe"]
