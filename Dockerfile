FROM golang:latest as builder

LABEL maintainer="Kondukto <dev@kondukto.io>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o kdt .

## Start a new stage from scratch
FROM alpine:latest 

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/kdt .

# Command to run the executable
ENTRYPOINT ["./kdt"]
