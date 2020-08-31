FROM golang:1.15

# the latest upstream release does not support go-1.15, build from source:
ARG GOLANGCI_TAG=8e63112ea3a41133967ac0e7007f3e6da1d178bd
RUN mkdir /src
WORKDIR /src
RUN curl -L https://github.com/golangci/golangci-lint/archive/${GOLANGCI_TAG}.tar.gz | \
  tar --strip-components=1 -xvvzf - && \
  go build -o /golangci-lint ./cmd/golangci-lint

ENTRYPOINT ["/golangci-lint", "run",\
            "--modules-download-mode=vendor",\
            "--deadline=15m"]
