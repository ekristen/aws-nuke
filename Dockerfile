# syntax=docker/dockerfile:1.13-labs
FROM alpine:3.21.2 AS base
RUN apk add --no-cache ca-certificates
RUN adduser -D aws-nuke

FROM ghcr.io/acorn-io/images-mirror/golang:1.21 AS build
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
WORKDIR /app
COPY --from=build --chmod=755 /src/bin/aws-nuke /usr/local/bin/aws-nuke
COPY --chmod=755 ./app/* /app/
RUN chmod 755 /usr/local/bin/aws-nuke /app/entrypoint.sh

# Install AWS CLI and jq using a virtual environment
USER root
RUN chown -R aws-nuke:aws-nuke /app
RUN apk add --no-cache curl jq aws-cli
USER aws-nuke

# Use the entry script as the command to run
CMD ["/app/entrypoint.sh"]
