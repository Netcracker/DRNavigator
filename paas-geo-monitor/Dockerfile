FROM artifactorycn.netcracker.com:17114/library/golang:1.18.3-alpine3.16 as builder

ENV GOPROXY=https://artifactorycn.netcracker.com/pd.sandbox-staging.go.group \
    GOSUMDB=off

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
COPY main.go ./
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o paas-geo-monitor ./main.go


FROM artifactorycn.netcracker.com:17114/library/alpine:3.16.1

ENV USER_UID=1001 USER_NAME=paas-geo-monitor GROUP_NAME=paas-geo-monitor

WORKDIR /
COPY --from=builder --chown=${USER_UID} workspace/paas-geo-monitor .
RUN addgroup ${GROUP_NAME} && adduser -D -G ${GROUP_NAME} -u ${USER_UID} ${USER_NAME}
USER ${USER_UID}

ENTRYPOINT  ["./paas-geo-monitor"]
