FROM golang:1.16.7 AS builder

WORKDIR /app
COPY go.mod /app
COPY go.sum /app
RUN go mod download

COPY . /app
RUN go build -o /update-go /app

FROM debian:buster-slim
COPY --from=builder /update-go /update-go
ENTRYPOINT ["/update-go"]
