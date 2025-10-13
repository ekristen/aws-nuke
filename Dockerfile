# syntax=docker/dockerfile:1.19-labs@sha256:dce1c693ef318bca08c964ba3122ae6248e45a1b96d65c4563c8dc6fe80349a2
FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1 as base
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
ENTRYPOINT ["/usr/local/bin/aws-nuke"]
COPY aws-nuke /usr/local/bin/aws-nuke
USER aws-nuke

FROM base
ENTRYPOINT ["/usr/local/bin/aws-nuke"]
COPY --from=build --chmod=755 /src/bin/aws-nuke /usr/local/bin/aws-nuke
RUN chmod +x /usr/local/bin/aws-nuke
USER aws-nuke