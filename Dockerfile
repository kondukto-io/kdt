FROM alpine:latest 

LABEL maintainer="Kondukto <dev@kondukto.io>"

# Create a group and user
RUN apk --no-cache add ca-certificates && addgroup -S appgroup && adduser -S appuser -G appgroup

# Tell docker that all future commands should run as the appuser user
USER appuser

# Change workdir to "/app"
WORKDIR /app

# Copy compiled binary to WORKDIR
COPY _release/kdt-linux-amd64 /app/kdt

# Command to run the executable
ENTRYPOINT ["./kdt"]
