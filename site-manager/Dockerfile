FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine3.20 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Add code generators
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.3
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0

# Copy the go source
COPY api/ api/
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/
COPY logger/ logger/
COPY config/ config/
COPY main.go ./

# Generate code
RUN go generate ./...

# Build
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on go build -a -o site-manager ./main.go


FROM --platform=$BUILDPLATFORM alpine:3.20.6

ENV USER_UID=1001 USER_NAME=site-manager GROUP_NAME=site-manager

WORKDIR /
COPY --from=builder --chown=${USER_UID} workspace/site-manager .

RUN apk update && \
    apk upgrade && \
    apk add bash && \
    addgroup ${GROUP_NAME} && adduser -D -G ${GROUP_NAME} -u ${USER_UID} ${USER_NAME}
USER ${USER_UID}

ENTRYPOINT  ["./site-manager"]
