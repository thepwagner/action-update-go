FROM debian:buster-slim

# These are special, and will be updated:
ARG CONTAINERD_VERSION=1.4.0
ARG CONTAINERD_SHASUM=1401ff0b102f15f499598ceeb95f10ee37fa13a7c7ab57a2c764472665d55860

# This will not be updated:
ARG CONTAINERD_URL=https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz
RUN apt-get update \
  && apt-get install -y ca-certificates curl \
  && curl -Lso /tmp/containerd.tgz $CONTAINERD_URL \
  && (echo "$CONTAINERD_SHASUM  /tmp/containerd.tgz" | sha256sum -c) \
  && tar -xvvf /tmp/containerd.tgz -C /usr/local/bin --strip-components=1 \
  && apt-get remove -y --purge curl \
  && apt-get autoremove -y --purge \
  && apt-get clean \
  && rm -Rf /var/lib/apt/lists/* /tmp/*
