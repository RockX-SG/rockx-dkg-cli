# RockX DKG CLI

- [RockX DKG CLI](#rockx-dkg-cli)
  - [Overview](#overview)
  - [DKG CLI](#dkg-cli)
    - [Pre-built binary](#pre-built-binary)
      - [Downloads](#downloads)
      - [Installation](#installation)
    - [Build from source](#build-from-source)
    - [Environment Variables](#environment-variables)
    - [Usage](#usage)
      - [Keygen](#keygen)
      - [Viewing results](#viewing-results)
      - [Generate deposit data](#generate-deposit-data)
      - [Get Keyshares](#get-keyshares)
  - [DKG Node](#dkg-node)
    - [Run using docker container](#run-using-docker-container)
      - [Environment Variables](#environment-variables-1)
      - [Running the container](#running-the-container)
  - [DKG Messenger](#dkg-messenger)
      - [Run using docker container](#run-using-docker-container-1)
  - [Examples](#examples)

## Overview

This CLI utility helps user to perform DKG based on FROST (Flexible Round-Optimized Schnorr Threshold) signature scheme. Unlike single-party signatures, threshold signatures involve multiple participants, each possessing a private key share. FROST reduces network overhead and supports efficient signing with two-round or single-round options while retaining genuine threshold signing capabilities. It can identify and exclude misbehaving participants, making it well-suited for practical MPC deployments. FROST ensures security against chosen-message attacks, assuming the hardness of the discrete logarithm problem and control by adversaries over fewer participants than the threshold.

This repository contains three primary services. SSV operators will execute the DKG Node service, while users creating validator keys on their local machines will utilize the CLI tool. Detailed instructions for building and running both the DKG Node and CLI services are provided in subsequent sections. Additionally, the repository includes a messenger service, serving as a communication layer connecting all DKG nodes for message exchange. RockX will host the messenger service for DKG usage in the SSV platform. If you prefer to run your own messenger service locally or as a Docker container, instructions for doing so are also available in a later section. Here's a brief overview of each service:

| Service Name | Description | 
| ------------ | ----------- |
| `rockx-dkg-cli`| A CLI utility for users to initiate DKG ceremony for keygen/resharing and generate files for DKG result, Deposit Data and SSV Keyshares |
| `rockx-dkg-node`| This is the core service run by each operator to participate in DKG ceremony |
| `rockx-dkg-messenger`| This is the communication layer enabling DKG Nodes to broadcase protocol messages to each other|

The relationship between these services can be summarized in the following diagram

![](/services_relationship.png)

## DKG CLI

### Pre-built binary

#### Downloads
|Version|Link| os|arch|
|-------|----|---|----|
|0.2.8| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.8/rockx-dkg-cli.0.2.8.darwin.arm64.tar.gz | darwin| arm64|
|0.2.8| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.8/rockx-dkg-cli.0.2.8.linux.amd64.tar.gz | linux| amd64|

#### Installation
1. Download the latest version of the cli tool from above links as per your system. The command for linux with amd64 architecture will be as follows:

```
wget https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.8/rockx-dkg-cli.0.2.8.linux.amd64.tar.gz
```
2. Extract the CLI tool

```
tar -xzvf rockx-dkg-cli.0.2.8.linux.amd64.tar.gz
```
3. Move the downloaded binary to your PATH

```
cp ./rockx-dkg-cli /usr/local/bin
```
> use sudo to run this command as root if you get permission denied error

4. Configure Messenger Service Endpoint

The default messenger address is preconfigured as https://dkg-messenger.rockx.com. If you're using this tool within the SSV platform, you can simply proceed without modifying the following environment variable. However, if you are hosting your own instance of the messenger service or running it locally on your machine (localhost), you have the option to set the following variable to specify the correct address for the messenger service configuration.
```
export MESSENGER_SRV_ADDR=https://dkg-messenger.rockx.com
```

5. Configure logging directory

The debug logs will be stored in the current directory. You can configure a custom path for logging by setting the following environment variable. Please note that you specify a location where this service has permission to create and write to a file.

```
export DKG_LOG_PATH=.
```

### Build from source

**Prerequisites**
1. Go 1.19
2. Docker (version 20 or later)
3. Docker Comose (1.29 or later)

To install the cli tool from source, clone this repository and run the following command

```
make build
```

The cli binary will be created at `./build/bin/rockx-dkg-cli`. You can add it to your PATH to access it directly by running

```
cd ./build/bin
export PATH=$PATH:`pwd`
```

### Environment Variables

| Variable | Description | Default Value |
| -------- | ----------- | ------------- |
| MESSENGER_SRV_ADDR | Address of messenger service | https://dkg-messenger.rockx.com
| USE_HARDCODED_OPERATORS | Use hardcoded private keys for operators for local testing. By default it's set to false and you need to set it to `true` to run DKG locally  | false

If you are running the example set of services (see [Examples](#example)) locally including the messenger service then make sure to set the following env variables

```
export MESSENGER_SRV_ADDR=http://0.0.0.0:3000
export USE_HARDCODED_OPERATORS=true
```

### Usage

By now you must have installed the DKG CLI tool. You can run following commands:

#### Keygen

The `keygen` command initiates keygen protocol. It takes the following parameters:

1.  --operator: Key value pair of operatorID (int) and operator's DKG node endpoint
2. --threshold: the minimum number of operators required to sign a message.
3. --withdrawal-credentials: The withdrawal credential associated with the validator
4. --fork-version: ETH fork version (for eg: prater)

Example: 

```
rockx-dkg-cli keygen \
    --operator 347="http://34.142.183.114:8081" \
    --operator 348="http://34.142.183.114:8080" \
    --operator 350="http://35.198.251.30:8080" \
    --operator 351="http://35.187.235.146:8080" \
    --threshold 3 \
    --withdrawal-credentials "0100000000000000000000001d2f14d2dffee594b4093d42e4bc1b0ea55e8aa7" \
    --fork-version "prater"
```

The CLI will return a request ID in the following format:
```
keygen init request sent with ID: 33a5b7fe2b415673c4d971e6c0b002ce7d583b6621dffb31
```

#### Viewing results

This command generates results of keygen/reshare by using the request ID generated in keygen/reshare command. It takes the following parameter:

1. --request-id: ID generated from calling keygen or reshare

Example:
```
rockx-dkg-cli get-dkg-results --request-id 33a5b7fe2b415673c4d971e6c0b002ce7d583b6621dffb31
```
This will write the results of the key generation/resharing process with the given request ID to a file of format `dkg_results_<request_id>_<timestamp>.json`

```
writing results to file: dkg_results_33a5b7fe2b415673c4d971e6c0b002ce7d583b6621dffb31_1701317995.json
```

#### Generate deposit data

Once the keygen is finished, you can also generate a Deposit Data file for depositing 32 ETH into your validator. It takes the following parameters:

1. --request-id: ID generated from calling keygen or reshare
2. --withdrawal-credentials: The withdrawal credential associated with the validator
4. --fork-version: ETH fork version (for eg: prater)

Example:

```
rockx-dkg-cli generate-deposit-data --request-id 33a5b7fe2b415673c4d971e6c0b002ce7d583b6621dffb31 -withdrawal-credentials "0100000000000000000000001d2f14d2dffee594b4093d42e4bc1b0ea55e8aa7" --fork-version "prater"
```
This will right the results to a json file in the following way

```
writing deposit data json to file deposit-data_1701318343.json
```

#### Get Keyshares

To distribute validator on SSV platform, you will need to select split key offine and then upload a keyshares file. To generate that keyshares file you can run the `get-keyshares` command. It takes the following parameters:

1. --request-id: ID generated from calling keygen or reshare
2. --operator: Key value pair of operatorID (int) and operator's DKG node endpoint
3. --owner-address: The cluster owner address (in the SSV contract)
4. --owner-nonce: The validator registration nonce of the account (owner address) within the SSV contract (increments after each validator registration), obtained using the ssv-scanner tool. (default: 0)

Example:
```
rockx-dkg-cli get-keyshares \
    --request-id 33a5b7fe2b415673c4d971e6c0b002ce7d583b6621dffb31 \
    --operator 347="http://34.142.183.114:8081" \
    --operator 348="http://34.142.183.114:8080" \
    --operator 350="http://35.198.251.30:8080" \
    --operator 351="http://35.187.235.146:8080" --owner-address "0x1d2f14d2dffee594b4093d42e4bc1b0ea55e8aa7" \
    --owner-nonce 0
```

It will generate a keyshares file in the following format that can be used to registor validator on SSV platform.

```
writing keyshares to file: keyshares-1701319254.json
```

## DKG Node

### Run using docker container

To run DKG node from a docker image, first you need to prepare application environment variable file and operator keys.

#### Environment Variables

| Variable | Description | requried/default value |
| -------- | ----------- | ---------------------- |
| NODE_OPERATOR_ID | SSV operator ID for this node | required |
| NODE_ADDR | Http address of the service | 0.0.0.0:8080 |
| NODE_BROADCAST_ADDR | The public ip or address of this DKG node | required |
| MESSENGER_SRV_ADDR | address of the messenger service | https://dkg-messenger.rockx.com |
| USE_HARDCODED_OPERATORS | use `true` for running local example | false |
| OPERATOR_PRIVATE_KEY | The raw base64 encoded RSA private key | Use either raw RSA private or JSON encode private key |
| OPERATOR_PRIVATE_KEY_PASSWORD_PATH | password file path for json encoded RSA private key | required |
| OPERATOR_PRIVATE_KEY_PATH | file path for json encoded RSA private key | required |

> Note: if your operator is configured using raw private key then use OPERATOR_PRIVATE_KEY. If it is configured using JSON encoded key then use OPERATOR_PRIVATE_KEY_PASSWORD_PATH and OPERATOR_PRIVATE_KEY_PATH

#### Running the container

By now you must have prepared env file and operator keys if you are using JSON encoded.

This might look something like this if you're using JSON encoded private key:

```
$ ls -Rw 1
.:
351.env
keys

./keys:
encryption_private_key.json
password
```

In case of raw private key you will just have the environment file

The environment file should look like these:

with raw private key:
```
NODE_OPERATOR_ID=351
NODE_ADDR=0.0.0.0:8080
NODE_BROADCAST_ADDR=http://35.187.235.146:8080
MESSENGER_SRV_ADDR=https://dkg-messenger.rockx.com
OPERATOR_PRIVATE_KEY=LS0tLS1CRUd...FURSBLRVktLS0tLQo=
```

with JSON encoded private key:
```
NODE_OPERATOR_ID=351
NODE_ADDR=0.0.0.0:8080
NODE_BROADCAST_ADDR=http://35.187.235.146:8080
MESSENGER_SRV_ADDR=https://dkg-messenger.rockx.com
OPERATOR_PRIVATE_KEY_PATH=/keys/encryption_private_key.json
OPERATOR_PRIVATE_KEY_PASSWORD_PATH=/keys/password
```

**Pull the latest docker image**

You can now pull the latest docker image for DKG node

```
docker pull asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-node:latest
```

**Run the container**

```
docker run -d --restart unless-stopped --name operator-351 -v /home/ubuntu/dkg/operator-351/keys:/keys --env-file /home/ubuntu/dkg/operator-351/351.env -p 8080:8080 asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-node:latest
```

You can check if your DKG node is correctly registered with the messenger service by running the following command and looking for your operatorID and your DKG node's public endpoint

```
curl -X GET https://dkg-messenger.rockx.com/topics/default
```

## DKG Messenger


At present, RockX provides a hosted messenger service at https://dkg-messenger.rockx.com. If you prefer to operate your own messenger service, you can do so by utilizing the Docker image and configuring it to map to your SSL-enabled public address or IP

#### Run using docker container

**Pull your docker image from GCP container registry**
```
docker pull asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-messenger:latest
```

**Run the container**
```
sudo docker run -d --name messenger -p 3000:3000 asia-southeast1-docker.pkg.dev/rockx-mpc-lab/rockx-dkg/rockx-dkg-messenger:latest
```

This will run the messenger service on port 3000

## Examples

The /env directory contains sample env files for 7 operator nodes with IDs from 1 to 7. You can run the following command to spin up 7 DKG nodes and a messenger node using following command

```
docker-compose up --build -d
```

You can then build the cli tool using `make build` command and test keygen and other commands locally. Make sure you set the following env vars to run locally

```
export MESSENGER_SRV_ADDR=http://0.0.0.0:3000
export USE_HARDCODED_OPERATORS=true
```


