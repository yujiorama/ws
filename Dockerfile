# first stage
FROM golang:1.12.5-stretch as build-app
ENV DEBIAN_FRONTEND=noninteractive
ENV XC_OS=linux
ENV XC_ARCH=amd64

WORKDIR /app
COPY . ./

RUN set -x \
 && apt-get update \
 && apt-get install -y --no-install-recommends build-essential

RUN set -x \
 && go mod download \
 && go build -o ws *.go

# final stage
FROM debian:stretch-slim as finish-app
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app
COPY --from=build-app /app/ws ./

RUN set -x \
 && useradd -d /app --no-create-home --shell /usr/sbin/nologin --uid 501 --user-group appuser \
 && chown -R appuser:appuser /app
USER appuser

ENTRYPOINT ["/app/ws"]
