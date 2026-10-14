ARG BASE_IMAGE=artifact.invicti.com/cgr-hardened/invicti.com/chainguard-base-fips:latest
FROM ${BASE_IMAGE}

LABEL maintainer="Invicti ASPM <support@invicti.com>"

# chainguard-base-fips runs as root by default; drop to the unprivileged chainguard user
USER 65532

WORKDIR /app

# Copy the FIPS 140-3 build (make fips) to WORKDIR
COPY --chown=65532:65532 --chmod=755 _release/kdt-linux-amd64-fips /app/kdt

# Command to run the executable
ENTRYPOINT ["./kdt"]
