FROM alpine:3.11 AS buf
ARG BUF_VERSION=v0.23.0
ARG BUF_CHECKSUM=096209b09e6c0e8b4e4c78b6fa40bad123f9d456cdb0b954e3939365f405ba2e

ARG BUFF_URL=https://github.com/bufbuild/buf/releases/download/${BUF_VERSION}/buf-Linux-x86_64
RUN apk --no-cache add --virtual .build curl \
  && curl -o /usr/local/bin/buf -L "$BUFF_URL" \
  && echo "$BUF_CHECKSUM  /usr/local/bin/buf" | sha256sum -c \
  && chmod +x /usr/local/bin/buf \
  && apk del .build