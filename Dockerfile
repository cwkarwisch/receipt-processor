FROM golang:1.23

# Set destination for COPY
WORKDIR /app

# Download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY *.go ./

# Build the binary
RUN go build -o /receipt-processor

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["/receipt-processor"]
