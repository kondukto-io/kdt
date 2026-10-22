ARG BASE_IMAGE=artifact.invicti.com/cgr-hardened/invicti.com/chainguard-base-fips:latest@sha256:a3d9b93074cee23617c80b573a0aeef4c12f57cd3cbec1d66f5c20da13d03ce0
FROM ${BASE_IMAGE}

LABEL maintainer="Invicti ASPM <support@invicti.com>"

# chainguard-base-fips runs as root by default; drop to the unprivileged chainguard user
USER 65532

WORKDIR /app

# Copy the FIPS 140-3 build of the linux/amd64 binary to WORKDIR
COPY --chown=65532:65532 --chmod=755 _release/kdt-linux-amd64 /app/kdt

# Command to run the executable
ENTRYPOINT ["./kdt"]
