FROM --platform=$BUILDPLATFORM debian:bookworm-slim@sha256:e67eeead736ff7bbe258ef088bff683635c561b1fcfefad0806c1df4ab49ce0e

RUN apt update && apt install -y netcat-openbsd iproute2 curl unzip && \
    curl -OL https://github.com/Netcracker/KubeMarine/releases/latest/download/package.zip && \
    unzip package.zip "dist/kubemarine-*.tar.gz" && \
    tar --wildcards -xv -f dist/kubemarine-*.tar.gz -C dist '*/ipip_check.gz'  && \
    find dist -name 'ipip_check.gz' -exec mv {} /bin \; && \
    gzip -d /bin/ipip_check.gz && \
    rm  package.zip && rm -R dist && \
    apt clean autoclean
