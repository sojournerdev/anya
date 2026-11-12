# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.27.4
ARG IMAGE_REVISION=unknown
ARG IMAGE_SOURCE=https://github.com/sojournerdev/anya
ARG IMAGE_VERSION=unknown
ARG IMAGE_CREATED=unknown

FROM golang:${GO_VERSION}-bookworm@sha256:4fa7c968ba93f1635670522dbb01a2c78150443e1385746924566079f36570ea AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download all

COPY . .

RUN test -n "$TARGETOS" && test -n "$TARGETARCH" \
    || (echo "TARGETOS and TARGETARCH are required; build with --platform" >&2; exit 1)

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 \
    GOOS="$TARGETOS" \
    GOARCH="$TARGETARCH" \
    go build \
      -trimpath \
      -buildvcs=false \
      -ldflags='-s -w -buildid=' \
      -o /out/node-agent \
      ./cmd/node-agent

FROM gcr.io/distroless/base-debian12:nonroot@sha256:7f0c72cd138b442ae0deeb69c08b1acf5525439ba251a49ad93c320a061567e5

ARG IMAGE_REVISION
ARG IMAGE_SOURCE
ARG IMAGE_VERSION
ARG IMAGE_CREATED

LABEL org.opencontainers.image.title="Anya" \
      org.opencontainers.image.description="Publishes validated NVIDIA GPU state over HTTP for GPU-aware workloads" \
      org.opencontainers.image.url="https://github.com/sojournerdev/anya" \
      org.opencontainers.image.source=$IMAGE_SOURCE \
      org.opencontainers.image.documentation="https://github.com/sojournerdev/anya/tree/main/docs" \
      org.opencontainers.image.version=$IMAGE_VERSION \
      org.opencontainers.image.revision=$IMAGE_REVISION \
      org.opencontainers.image.created=$IMAGE_CREATED \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.authors="sojournerdev" \
      org.opencontainers.image.vendor="sojournerdev"

COPY --from=build /out/node-agent /node-agent

USER nonroot:nonroot
STOPSIGNAL SIGTERM
ENTRYPOINT ["/node-agent"]
