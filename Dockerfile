# syntax=docker/dockerfile:1.3-labs
FROM alpine:3.19.0 as base
RUN apk add --no-cache ca-certificates
RUN adduser -D aws-nuke

FROM ghcr.io/acorn-io/images-mirror/golang:1.21 AS build
COPY / /src
WORKDIR /src
RUN \
  --mount=type=cache,target=/go/pkg \
  --mount=type=cache,target=/root/.cache/go-build \
  go build -o bin/aws-nuke main.go

FROM base AS goreleaser
ENTRYPOINT ["/usr/local/bin/aws-nuke"]
COPY aws-nuke /usr/local/bin/aws-nuke
USER aws-nuke

FROM base
ENTRYPOINT ["/usr/local/bin/aws-nuke"]
COPY --from=build /src/bin/aws-nuke /usr/local/bin/aws-nuke
USER aws-nuke