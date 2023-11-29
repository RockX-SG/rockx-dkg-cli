# RockX DKG CLI

- [RockX DKG CLI](#rockx-dkg-cli)
  - [Overview](#overview)
  - [DKG CLI](#dkg-cli)
    - [Pre-built binary](#pre-built-binary)
      - [Downloads](#downloads)
      - [Installation](#installation)
    - [Build from source](#build-from-source)

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
|0.2.7| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.7/rockx-dkg-cli.0.2.7.darwin.arm64.tar.gz | darwin| arm64|
|0.2.7| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.7/rockx-dkg-cli.0.2.7.linux.amd64.tar.gz | linux| amd64|

#### Installation
1. Download the latest version of the cli tool from above links as per your system. The command for linux with amd64 architecture will be as follows:

```
wget https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.7/rockx-dkg-cli.0.2.7.linux.amd64.tar.gz
```
2. Extract the CLI tool

```
tar -xzvf rockx-dkg-cli.0.2.7.linux.amd64.tar.gz
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
6. 

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
If you are running the example set of services (see [Examples](#example)) locally including the messenger service then make sure to set the following env variables

```
export MESSENGER_SRV_ADDR=http://0.0.0.0:3000
export USE_HARDCODED_OPERATORS=true
```

We employ Docker Compose to establish seven DKG nodes running locally, spanning from localhost:8081 to localhost:8087. Simultaneously, we set up a messenger node locally at localhost:3000. It's important to note that this configuration does not establish a connection to SSV's Operator Registry and utilizes a predefined set of operator keys hardcoded into the setup. The key shares generated through this configuration are solely intended for demonstration purposes. his configuration is designed for your local exploration and debugging of the DKG solution. In a production environment, there is no need to set the `USE_HARDCODED_OPERATORS`  environment variable as it is already set to `false`` by default.