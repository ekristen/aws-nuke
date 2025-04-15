# syntax=docker/dockerfile:1.15-labs@sha256:8378c88c56f2a6c038705487ce1e447c61c48557cd6a76aea4d53e255304260a
FROM alpine:3.21.3@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c as base
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