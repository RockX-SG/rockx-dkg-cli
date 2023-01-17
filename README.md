# FROST DKG (Distributed Key Generation) Demo

## Validator Onboarding

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

## FROST DKG

This repository has a set of services that demonstrate how to use frost DKG to generate a validator public key and shares that are split between operators using Shamir Secret Sharing.
It includes:

1. An API service that provides ways to start keygen and check results, and retrieve deposit data
2. A messenger service that allows operators to register and handles messages between operators
3. A Node service that receives messages from other nodes or API service and runs the DKG algorithm to generate the validator public key
4. To use this repository, you need to install it, start the keygen process, view the results and retrieve the deposit data to be uploaded on the Ethereum 2.0 deposit contract.

## Getting Started
### Prerequisites
1. Go 1.19
2. Docker (20 or later)
3. Docker Compose (1.29 or later)

### Installation
This code repository contains a Docker Compose configuration file to set up and run all necessary services. To start these services, run the following command:

```
docker-compose up -d
```
### Key Generation
To trigger a key generation process with a threshold of t out of the 4 operator nodes, make a POST request to the /keygen endpoint with the following JSON payload:
```
curl --location --request POST 'http://0.0.0.0:8000/keygen' \
--header 'Content-Type: application/json' \
--data-raw '{
    "operators": {
        "1": "http://host.docker.internal:8081",
        "2": "http://host.docker.internal:8082",
        "3": "http://host.docker.internal:8083",
        "4": "http://host.docker.internal:8084"
    },
    "threshold": 3
}'
```
The API will return a request ID in the following format:
```
{
    "request_id": "5d2aea535cd70f5a18f78b83953444e2eeb5a978902edb21"
}
```
### Viewing Results
To view the results of a key generation process, use the request ID returned from the previous step and make a GET request to the /data/{request_id} endpoint:
```
curl --location --request GET 'http://0.0.0.0:8000/data/5d2aea535cd70f5a18f78b83953444e2eeb5a978902edb21'
```
This will return the results of the key generation process with the given request ID.

### Verifying Results
To verify results, use Verify tool with Validator Public Key and Deposit Data signature
```
# Build verify tool
make build_verify

# Run verify tool
# ./build/bin/verify <validator_public_key> <deposit_data_sig>

./build/bin/verify 92b0a3da8664b1aa8579267393477229e2d46dff2876f820f51a44e9cfd2aeafa61b54e33c4265a7f68c94b86fcce181 8eb4d37745365e0b9f782b05c578b20d1998c9fcb35bbc597d495ea840ea6e104be83ed81fe500134fed8e7f029e0c9f0096b441f82a7fe6695fbc449456751aa1a3625e97f983309637db348994f262e4f263b85eb463782b75a6b8ea0a054c

# Output
# ~ signature verification succeeded
```
### Get Deposit data
To download deposit data run the following command in the browser. It will download a json file with name of format `deposit-data_*.json`
```
http://0.0.0.0:8000/deposit_data/<request_id>'
# for eg. http://0.0.0.0:8000/deposit_data/b1e3e5d676b6829a7f9115f936e244a72fea9c448c7fcde3
```

The downloaded file can be verified at https://goerli.launchpad.ethereum.org/en/overview