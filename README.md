## FROST DKG (Distributed Key Generation) Demo

### Prerequisites
1. Go 1.19
2. Docker (20 or later)
3. Docker Compose (1.29 or later)

### Getting Started
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
