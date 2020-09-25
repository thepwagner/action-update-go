FROM debian:buster-slim

# These are special, and will be updated:
ARG ELIXIR_VERSION=v1.10.3
ARG ELIXIR_CHECKSUM=fc6d06ad4cc596b2b6e4f01712f718200c69f3b9c49c7d3b787f9a67b36482658490cf01109b0b0842fc9d88a27f64a9aba817231498d99fa01fa99688263d55

# This will not be updated:
RUN apt-get update \
  && apt-get install -y ca-certificates curl gnupg unzip \
  && curl -Lso /tmp/erlang.deb https://packages.erlang-solutions.com/erlang-solutions_1.0_all.deb \
  && dpkg -i /tmp/erlang.deb \
  && apt-get update \
  && apt-get install -y esl-erlang \
  && curl -Lso /tmp/elixir.zip https://github.com/elixir-lang/elixir/releases/download/${ELIXIR_VERSION}/Precompiled.zip \
  && echo "$ELIXIR_CHECKSUM  /tmp/elixir.zip" | sha512sum -c - \
  && unzip -d /usr/local/elixir -x /tmp/elixir.zip \
  && apt-get remove -y --purge curl gnupg unzip \
  && apt-get autoremove -y --purge \
  && apt-get clean \
  && rm -Rf /var/lib/apt/lists/* /tmp/*
