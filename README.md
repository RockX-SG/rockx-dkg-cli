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
    "threshold": 3,
    "withdrawal_credentials": "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f",
    "fork_version": "prater"
}'
```
The API will return a request ID in the following format:
```
{
    "request_id": "59c971e3477e19b48fc467bb6e300d8eab34cf32ae7eba35"
}
```

### Resharing
To trigger a resharing process with a threshold of t out of the 4 operator nodes, make a POST request to the /resharing endpoint with the following JSON payload:
```
curl --location --request POST 'http://0.0.0.0:8000/resharing' \
--header 'Content-Type: application/json' \
--data-raw '{
    "operators": {
        "5": "http://host.docker.internal:8085",
        "6": "http://host.docker.internal:8086",
        "7": "http://host.docker.internal:8087",
        "8": "http://host.docker.internal:8088"
    },
    "threshold": 3,
    "validator_pk": "8eb0f05adc697cdcbdf8848f7f1e8c2277f4fc7b0efc97ceb87ce75286e4328db7259fc0c1b39ced0c594855a30d415c",
    "operators_old": {
        "1": "http://host.docker.internal:8081",
        "2": "http://host.docker.internal:8082",
        "3": "http://host.docker.internal:8083"
    }
}'
```
The API will return a request ID in the following format:

```json
{
    "request_id": "59c971e3477e19b48fc467bb6e300d8eab34cf32ae7eba35"
}
```

### Viewing Results
To view the results of a key generation process (or resharing), use the request ID returned from the previous step and make a GET request to the /data/{request_id} endpoint:
```
curl --location --request GET 'http://0.0.0.0:8000/data/59c971e3477e19b48fc467bb6e300d8eab34cf32ae7eba35'
```
This will return the results of the key generation process with the given request ID.

### Verifying Results
To verify results, use Verify tool with Validator Public Key and Deposit Data signature
```
# Build verify tool
make build_verify

# Run verify tool
# ./build/bin/verify <fork_version> <validator_public_key> <deposit_data_sig> <withdrawal credentials>

./build/bin/verify prater 87d7a269ec845bd363fd2c6b2e8e61d5314725d5456ca5c4c8397d33d3052bb2c641e50ee78939f9deed429dff4f48ad 8ea5d0dddec9aa797fbb624c5732ea47fea89cc63adb391e15892e7b849a86edc93de80bace9cc06d85243d92c718fbb0c2cef9a8f5dd61f7af534ff1c211966fa581605410ea5bc13848a52626a612d690d5f8aabc80c0b619be2ef785ed88d 010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f

# Output
# ~ signature verification succeeded
```
### Get Deposit data
To download deposit data run the following command in the browser. It will download a json file with name of format `deposit-data_*.json`
```
http://0.0.0.0:8000/deposit_data/<request_id>'
# for eg. http://0.0.0.0:8000/deposit_data/59c971e3477e19b48fc467bb6e300d8eab34cf32ae7eba35
```

The downloaded file can be verified at https://goerli.launchpad.ethereum.org/en/overview
