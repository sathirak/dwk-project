# Build stage
FROM golang:1.23.4-alpine AS builder

WORKDIR /build

# Copy only dependency files first
COPY go.mod ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:3.18

LABEL maintainer="Project Maintainer"
LABEL description="Todo Application"

WORKDIR /app

# Create shared directory
RUN mkdir -p /usr/src/app/shared

# Copy binary from builder
COPY --from=builder /build/main .

# Use exec form with full path
ENTRYPOINT ["/app/main"]