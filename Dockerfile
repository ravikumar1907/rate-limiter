# Use an official Golang image
FROM golang:1.19-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy everything from the current directory to the container
COPY . .

# Install go dependencies
RUN go mod download

# Build the Go app
RUN go build -o rate-limiter .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./rate-limiter"]
