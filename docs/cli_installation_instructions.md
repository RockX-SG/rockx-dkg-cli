## RockX DKG CLI - Installation instructions
---

Downloads
|Version|Link| os|arch|
|-------|----|---|----|
|0.1.1| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/0.1.1/rockx-dkg-cli.0.1.1.darwin.arm64.tar.gz| darwin| arm64|
|0.1.1| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/0.1.1/rockx-dkg-cli.0.1.1.linux.amd64.tar.gz| linux| amd64|


### Installation (for os: darwin and arch: arm64)

1. Download the latest version of the cli tool from above page

```
wget https://github.com/RockX-SG/rockx-dkg-cli/releases/download/0.1.1/rockx-dkg-cli.0.1.1.darwin.arm64.tar.gz
```

2. Extract the cli

```
tar -xzvf rockx-dkg-cli.0.1.1.darwin.arm64.tar.gz
```

3. Move the file to your PATH

```
cp ./rockx-dkg-cli /usr/local/bin
```

4. Perform DKG
```
rockx-dkg-cli keygen \
 --operator 1="http://34.143.199.161:8080" \
 --operator 2="http://35.240.226.66:8080" \
 --operator 3="http://34.87.9.120:8080" \
 --operator 4="http://34.124.174.255:8080" \
 --threshold 3 \
 --withdrawal-credentials "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f" \
 --fork-version "prater"
```

5. View Results
```
rockx-dkg-cli get-dkg-results \
 --request-id f99672b06987b3ae88a2f884488d684373bb18be8eb72e5d
```

6. Generate Deposit Data
```
rockx-dkg-cli generate-deposit-data \
 --withdrawal-credentials "010000000000000000000000535953b5a6040074948cf185eaa7d2abbd66808f" \
 --fork-version "prater" \
 --request-id f99672b06987b3ae88a2f884488d684373bb18be8eb72e5d
```