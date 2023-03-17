# DKG Operator Node - installation

### Dockerfile
[here](../build/docker/node/Dockerfile)
```
FROM     golang:1.19-buster AS builder
WORKDIR  /app
COPY     . .
RUN      go mod download && make build_node

FROM     ubuntu:18.04
WORKDIR  /app
VOLUME   /keys
COPY     --from=builder /app/build/bin/node /app/node
CMD      ["./node"] 
```

### Environment variables
```
NODE_OPERATOR_ID=1
NODE_ADDR=0.0.0.0:8080
NODE_BROADCAST_ADDR=<public ip or public address>
MESSENGER_SRV_ADDR=https://dkg-messenger.rockx.com
KEYSTORE_FILE_PATH=/keys/<keystore file name>
KEYSTORE_PASSWORD=password
USE_HARDCODED_OPERATORS=false
```

> Note: keep USE_HARDCODED_OPERATORS=false to use SSV operator registry instead of hardcoded values


### Docker command to run the containers

#### Keystore file

Create a folder like `keystorefiles` here and mv your keystore file. 

```
ls keystorefiles/
keystore-m_12381_3600_3_0_0-1677466776.json

```

#### Environment Variables file

Create a folder `env` and store your env file here

```
ls env/
operator.1.env
```

#### Pull your docker image from GCP container registry
```
docker pull asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-node:latest
```

#### Run the container with the env file and the keystore folder
```
docker run -d --name operator-node -v ./keystorefiles:/keys --env-file ./env/operator.1.env -p 8080:8080 asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-node
```


