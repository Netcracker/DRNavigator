FROM debian:bookworm-slim

RUN apt update && \
    apt install -y netcat-openbsd iproute2 curl unzip && \
    curl -OL https://github.com/Netcracker/KubeMarine/releases/latest/download/package.zip && \
    unzip package.zip "dist/kubemarine-*.tar.gz" && \
    tar --wildcards -xO -f dist/kubemarine-*.tar.gz '*/ipip_check.gz'| gzip -d >bin/ipip_check && \
    chmod +x /bin/ipip_check && \
    rm  package.zip && rm -R dist && \
    apt clean autoclean
