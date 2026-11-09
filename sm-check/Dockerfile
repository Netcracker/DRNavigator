FROM --platform=$BUILDPLATFORM debian:bookworm-slim@sha256:734a72b3d6142a71993b4bc580ca01d8ca687f77a2c422d96a961d3376b179bc

RUN apt update && apt install -y netcat-openbsd iproute2 curl unzip && \
    curl -OL https://github.com/Netcracker/KubeMarine/releases/latest/download/package.zip && \
    unzip package.zip "dist/kubemarine-*.tar.gz" && \
    tar --wildcards -xv -f dist/kubemarine-*.tar.gz -C dist '*/ipip_check.gz'  && \
    find dist -name 'ipip_check.gz' -exec mv {} /bin \; && \
    gzip -d /bin/ipip_check.gz && \
    rm  package.zip && rm -R dist && \
    apt clean autoclean
