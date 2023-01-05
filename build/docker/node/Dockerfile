FROM     golang:1.19-buster AS builder
WORKDIR  /app
COPY     . .
RUN      go mod download && make build_node

FROM     ubuntu:18.04
WORKDIR  /app
EXPOSE   8080
COPY     --from=builder /app/build/bin/node /app/node
CMD      ["./node"] 