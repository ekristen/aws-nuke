# syntax=docker/dockerfile:1.22-labs@sha256:4c116b618ed48404d579b5467127b20986f2a6b29e4b9be2fee841f632db6a86
FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS base
RUN apk add --no-cache ca-certificates
RUN adduser -D aws-nuke

FROM ghcr.io/acorn-io/images-mirror/golang:1.21@sha256:856073656d1a517517792e6cdd2f7a5ef080d3ca2dff33e518c8412f140fdd2d AS build
COPY / /src
WORKDIR /src
ENV CGO_ENABLED=0
RUN \
  --mount=type=cache,target=/go/pkg \
  --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags '-s -w -extldflags="-static"' -o bin/aws-nuke main.go

FROM base AS goreleaser
ARG TARGETPLATFORM
ENTRYPOINT ["/usr/local/bin/aws-nuke"]
COPY $TARGETPLATFORM/aws-nuke /usr/local/bin/aws-nuke
USER aws-nuke

FROM base
ENTRYPOINT ["/usr/local/bin/aws-nuke"]
COPY --from=build --chmod=755 /src/bin/aws-nuke /usr/local/bin/aws-nuke
RUN chmod +x /usr/local/bin/aws-nuke
USER aws-nuke
