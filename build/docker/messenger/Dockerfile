FROM     golang:1.19-buster AS builder
WORKDIR  /app
COPY     . .
RUN      go mod download && make build_messenger

FROM     ubuntu:18.04
WORKDIR  /app
EXPOSE   3000
COPY     --from=builder /app/build/bin/messenger /app/messenger
CMD      ["./messenger"] 