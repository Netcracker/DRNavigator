FROM --platform=$BUILDPLATFORM golang:1.22.6-alpine3.20 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY logger/ logger/
COPY main.go ./
# Build
# GOMEMLINIT is 90% of kuber limits for garbage-collector
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o paas-geo-monitor ./main.go

FROM --platform=$BUILDPLATFORM alpine:3.19.1

ENV USER_UID=1001 USER_NAME=paas-geo-monitor GROUP_NAME=paas-geo-monitor

RUN apk --no-cache add curl=8.9.0-r0

WORKDIR /
COPY --from=builder --chown=${USER_UID} workspace/paas-geo-monitor .
RUN apk update && \
    apk upgrade && \
    apk add bash && \
    addgroup ${GROUP_NAME} && adduser -D -G ${GROUP_NAME} -u ${USER_UID} ${USER_NAME}
USER ${USER_UID}

ENTRYPOINT  ["./paas-geo-monitor"]
