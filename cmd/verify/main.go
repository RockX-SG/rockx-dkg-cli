package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/bloxapp/ssv-spec/types"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	types.InitBLS()
}

func main() {
	fork := types.NetworkFromString(os.Args[1])
	pkbytes, _ := hex.DecodeString(os.Args[2])
	depositSig := os.Args[3]
	withdrawalCredentials, _ := hex.DecodeString(os.Args[4])

	signingRoot, _, err := types.GenerateETHDepositData(pkbytes, withdrawalCredentials, fork.ForkVersion(), types.DomainDeposit)
	checkErr(err)

	var (
		pk  bls.PublicKey
		sig bls.Sign
	)

	err = pk.Deserialize(pkbytes)
	checkErr(err)

	err = sig.DeserializeHexStr(depositSig)
	checkErr(err)

	if sig.VerifyByte(&pk, signingRoot) {
		fmt.Println("signature verification succeeded")
	} else {
		panic("signature verification failed")
	}
}
