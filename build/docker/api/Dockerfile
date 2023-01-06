FROM     golang:1.19-buster AS builder
WORKDIR  /app
COPY     . .
RUN      go mod download && make build

FROM     ubuntu:18.04
WORKDIR  /app
COPY     --from=builder /app/build/bin/api /app/api
CMD      ["./api"] 