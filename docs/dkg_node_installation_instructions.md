# DKG Operator Node - installation

### Environment variables for .env file
```
NODE_OPERATOR_ID=1
NODE_ADDR=0.0.0.0:8080
NODE_BROADCAST_ADDR=<public ip or public address>
MESSENGER_SRV_ADDR=https://dkg-messenger.rockx.com
OPERATOR_PRIVATE_KEY=<SSV_OPERATOR_PRIVATE_KEY>
USE_HARDCODED_OPERATORS=false
```

> Note: keep USE_HARDCODED_OPERATORS=false to use SSV operator registry instead of hardcoded values


### Docker command to run the containers

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

#### Run the container with the env file
```
docker run -d --name operator-node --env-file ./env/operator.1.env -p 8080:8080 asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-node
```