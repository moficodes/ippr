# Use the official Go image as a builder
FROM golang:1.25.3-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ippr .

# Use a minimal base image for the final image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/ippr .

# Copy the templates directory
COPY --from=builder /app/templates ./templates

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./ippr"]
