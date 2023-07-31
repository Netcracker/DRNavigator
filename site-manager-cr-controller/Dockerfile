FROM golang:1.20.2-alpine3.16 as builder

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
COPY config/ config/
COPY main.go ./
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o site-manager-cr-controller ./main.go


FROM alpine:3.16.2

ENV USER_UID=1001 USER_NAME=site-manager-cr-controller GROUP_NAME=site-manager-cr-controller

WORKDIR /
COPY --from=builder --chown=${USER_UID} workspace/site-manager-cr-controller .
RUN addgroup ${GROUP_NAME} && adduser -D -G ${GROUP_NAME} -u ${USER_UID} ${USER_NAME}
USER ${USER_UID}

ENTRYPOINT  ["./site-manager-cr-controller"]
