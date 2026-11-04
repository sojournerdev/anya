# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.27.3
ARG IMAGE_REVISION=unknown
ARG IMAGE_SOURCE=https://github.com/sojournerdev/anya
ARG IMAGE_VERSION=unknown
ARG IMAGE_CREATED=unknown

FROM golang:${GO_VERSION}-bookworm@sha256:dc0963f32053e4bc451d24d87e3892d42d01021b61f093e98eb95c0352bb84b0 AS build

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

LABEL org.opencontainers.image.source=$IMAGE_SOURCE \
      org.opencontainers.image.revision=$IMAGE_REVISION \
      org.opencontainers.image.version=$IMAGE_VERSION \
      org.opencontainers.image.created=$IMAGE_CREATED

COPY --from=build /out/node-agent /node-agent

USER nonroot:nonroot
STOPSIGNAL SIGTERM
ENTRYPOINT ["/node-agent"]
