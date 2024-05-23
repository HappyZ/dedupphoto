# Start from a Golang image to compile the application
FROM golang:1.22 AS builder

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy the source code into the container
COPY . .

# Build the Go app statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /tmp/dedupphoto .

# Start a new stage from scratch
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /tmp/dedupphoto /dedupphoto

# Command to run the executable with arguments
CMD ["/dedupphoto", "--folder", "/myfolder", "--dryrun=false", "--trashbin", "/mytrashbin"]