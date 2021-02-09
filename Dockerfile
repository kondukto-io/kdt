FROM alpine:latest 

LABEL maintainer="Kondukto <dev@kondukto.io>"

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY _release/kdt-linux /app/kdt

# Command to run the executable
ENTRYPOINT ["./kdt"]
