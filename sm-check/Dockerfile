FROM debian:bookworm-slim

ENV VER 0.28.1

RUN apt update && \
    apt install -y netcat-openbsd iproute2 curl unzip && \
    curl -OL https://github.com/Netcracker/KubeMarine/releases/download/v${VER}/package.zip && \
    unzip package.zip && \
    tar -zxvf dist/kubemarine-${VER}.tar.gz && \  
    gunzip kubemarine-${VER}/kubemarine/resources/scripts/ipip_check.gz && \
    cp kubemarine-${VER}/kubemarine/resources/scripts/ipip_check /bin/ && \
    rm package.zip && rm -R kubemarine-${VER} && rm -R dist && \
    apt clean autoclean
