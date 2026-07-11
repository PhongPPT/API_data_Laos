FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api_server ./cmd/api/main.go

# Use a lightweight alpine image for the final container
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/api_server .
# Copy the swagger docs since they are served statically
COPY --from=builder /app/docs ./docs

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./api_server"]
