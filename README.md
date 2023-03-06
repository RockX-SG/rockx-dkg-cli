## RockX DKG CLI

### Overview - Validator Onboarding
![](/validator_onboarding.png)

To become a validator on Ethereum 2.0 using SSV, you need to:

1. Create a signing key and withdrawal key using instructions on the Ethereum Launchpad
2. Make a deposit of 32 ETH along with deposit data via the Ethereum 2.0 deposit contract
3. Select operators in SSV app
4. Split the signing key to the selected operators
5. Selected operators will begin performing their duties for the new validators

> Note: This process is without the use of Distributed Key Generation (DKG) method to generate the keys and deposit data.

With DKG, instead of creating signing key and deposit data manually, you can:

1. Send a request through the SSV app to generate the keys and deposit data automatically.
2. The request can be bundled with operator selection, so no need for separate steps.
3. The operator will perform the DKG protocol to generate the keys and deposit data
4. User needs to supply the withdrawal credential (hash of the withdrawal key)
5. Retrieve deposit data from SSV app after DKG and use it to deposit 32 ETH via Ethereum 2.0 deposit contract
6. Pre-selected operators will begin performing their duties for the new validators after deposit is made.

The process of creating keys and deposit data with DKG is different from without DKG, but they are compatible with each other. DKG gives users another option to generate keys and in some cases it may be required, but the existing way of generating keys will still work and no action is needed from the users.

## Getting Started

This repository has a set of services that demonstrate how to use frost DKG to generate a validator public key and shares that are split between operators using Shamir Secret Sharing.
It includes:

1. An CLI utility that provides ways to start keygen/resharing and get results to retrieve validator public key, and generate deposit data in json format
2. A messenger service that allows operators to register and handles messages between operators
3. A Node service that receives messages from other nodes or CLI tool and runs the DKG algorithm to generate the validator public key

### Prerequisites
1. Go 1.19
2. Docker (20 or later)
3. Docker Compose (1.29 or later)

### Installation
This code repository contains a Docker Compose configuration file to set up and run all necessary services. To start these services, run the following command:

```
docker-compose up -d
```

To install the cli tool, run the following command:
```shell
make build
```

The cli binary can be found at `./build/bin` as `rockx-dkg-cli`. You can add it to you PATH to access it directly buy running

```
cd ./build/bin
export PATH=$PATH:`pwd`
```

You can check all the available command by just typing `rockx-dkg-cli`
```
NAME:
   rockx-dkg-cli - A cli tool to run DKG for keygen and resharing and generate deposit data

USAGE:
   rockx-dkg-cli [global options] command [command options] [arguments...]

COMMANDS:
   keygen, k                   start keygen process
   resharing, r                start resharing process
   get-dkg-results, gr         get validator-pk and key shares of all operators
   generate-deposit-data, gdd  generate deposit data in json format
   help, h                     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

### Key Generation
The `keygen` command is used to generate a new set of key shares using the distributed key generation protocol. The command takes the following parameters:

##### Command Options
--operator: The key value pair of operatorID (int) and server addr of the dkg operator node 
--threshold: The minimum number of operators required to sign a message.
--withdrawal-credentials: The withdrawal credentials associated with the validator account.
--fork-version: The fork version value.

##### Example:
```
rockx-dkg-cli keygen --operator 1="http://0.0.0.0:8081" --operator 2="http://0.0.0.0:8082" --operator 3="http://0.0.0.0:8083" --operator 4="http://0.0.0.0:8084" --threshold 3 --withdrawal-credentials "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f" --fork-version "prater"
```

The CLI will return a request ID in the following format:
```
keygen init request sent with ID: a6e2cb702e163a328c0ab80b29a4d444feb3ac948088462f
```

### Resharing
The `resharing` command is used to reshare an existing validator public key from old committee members to new committee

##### Command Options
--operator: The key value pair of operatorID (int) and server addr of the dkg operator node in the new committee
--old-operator: The key value pair of operatorID (int) and server addr of the dkg operator node from the old committee. Atleast previous threshold number of operators are required to successfully perform resharing
--threshold: The minimum number of operators required to sign a message.
--validator-pk: The public key of the validator account.

##### Example:
```
rockx-dkg-cli resharing --operator 5="http://0.0.0.0:8085"  --operator 6="http://0.0.0.0:8086"  --operator 7="http://0.0.0.0:8087"  --operator 8="http://0.0.0.0:8088" --old-operator 1="http://0.0.0.0:8081" --old-operator 2="http://0.0.0.0:8082" --old-operator 3="http://0.0.0.0:8083"  --threshold 3 --validator-pk adf8b634f1c2bb64fe61af95b208a2a7bdac0d2d15963f83463bdb85c7e726250bfa3a390bf01edfc0700d61f4bee579
```
The CLI will return a request ID in the following format:

```
resharing init request sent with ID: c9e8c174060ee45bf86aaea3e409d8ee48a8fcb3d008fd18
```

### Viewing Results
To view the results of a key generation process (or resharing), use the request ID returned from the previous step and use `get-dkg-results` command

##### Command Options
--request-id: request id generated from calling keygen or resharing command

##### Example:
```
rockx-dkg-cli get-dkg-results --request-id c9e8c174060ee45bf86aaea3e409d8ee48a8fcb3d008fd18
```
This will write the results of the key generation/resharing process with the given request ID to a file of format `dkg_results_<request_id>_<timestamp>.json`

```
writing results to file: dkg_results_c9e8c174060ee45bf86aaea3e409d8ee48a8fcb3d008fd18_1678083260.json
```

### Generate Deposit data
To generate deposit data run the command `generate-deposit-data` from the cli. It will generate a json file with name format as `deposit-data_*.json`

##### Command Options
--request-id: request id of previously ran keygen process.
--withdrawal-credentials: The withdrawal credentials associated with the validator account.
--fork-version: The fork version value.

#### Example:
```
rockx-dkg-cli generate-deposit-data --withdrawal-credentials "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f" --fork-version "prater" --request-id a6e2cb702e163a328c0ab80b29a4d444feb3ac948088462f
```

The generated file can be verified at https://goerli.launchpad.ethereum.org/en/overview

### Verifying Results
To verify results, use Verify tool with Validator Public Key and Deposit Data signature
```
# Build verify tool
make build_verify

# Run verify tool
# ./build/bin/verify <fork_version> <validator_public_key> <deposit_data_sig> <withdrawal credentials>

./build/bin/verify prater 87d7a269ec845bd363fd2c6b2e8e61d5314725d5456ca5c4c8397d33d3052bb2c641e50ee78939f9deed429dff4f48ad 8ea5d0dddec9aa797fbb624c5732ea47fea89cc63adb391e15892e7b849a86edc93de80bace9cc06d85243d92c718fbb0c2cef9a8f5dd61f7af534ff1c211966fa581605410ea5bc13848a52626a612d690d5f8aabc80c0b619be2ef785ed88d 010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f

# Output
# signature verification succeeded
```

