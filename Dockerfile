# First stage: build the Go application
FROM golang:1.20.3 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files, and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webreroute .

# Second stage: create the final Docker image
FROM alpine:latest

# Set the working directory
WORKDIR /root

# Copy the binary from the builder stage
COPY --from=builder /app/webreroute .

# Expose the port used by the application
EXPOSE 8088

# Run the application
CMD ["./webreroute"]


