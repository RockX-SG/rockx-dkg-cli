## RockX DKG CLI - Installation instructions
---

Downloads
|Version|Link| os|arch|
|-------|----|---|----|
|0.2.3| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.3/rockx-dkg-cli.0.2.3.darwin.arm64.tar.gz | darwin| arm64|
|0.2.3| https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.3/rockx-dkg-cli.0.2.3.linux.amd64.tar.gz | linux| amd64|


### Installation (for os: darwin and arch: arm64)

1. Download the latest version of the cli tool from above page

```
wget https://github.com/RockX-SG/rockx-dkg-cli/releases/download/v0.2.3/rockx-dkg-cli.0.2.3.linux.amd64.tar.gz
```

2. Extract the cli

```
tar -xzvf rockx-dkg-cli.0.2.3.linux.amd64.tar.gz
```

3. Move the file to your PATH and Set messenger service address for the cli

```
cp ./rockx-dkg-cli /usr/local/bin
export MESSENGER_SRV_ADDR=https://dkg-messenger.rockx.com
```

4. Perform DKG
```
rockx-dkg-cli keygen \
 --operator 1="http://34.143.199.161:8080" \
 --operator 2="http://35.240.226.66:8080" \
 --operator 3="http://34.87.9.120:8080" \
 --operator 4="http://34.124.174.255:8080" \
 --threshold 3 \
 --withdrawal-credentials "0100000000000000000000001d2f14d2dffee594b4093d42e4bc1b0ea55e8aa7" \
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
 --withdrawal-credentials "0100000000000000000000001d2f14d2dffee594b4093d42e4bc1b0ea55e8aa7" \
 --fork-version "prater" \
 --request-id f99672b06987b3ae88a2f884488d684373bb18be8eb72e5d
```

7. Generate Keyshares
```
rockx-dkg-cli get-keyshares --request-id f99672b06987b3ae88a2f884488d684373bb18be8eb72e5d
```
