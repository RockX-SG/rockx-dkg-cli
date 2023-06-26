/*
 * ==================================================================
 *Copyright (C) 2022-2023 Altstake Technology Pte. Ltd. (RockX)
 *This file is part of rockx-dkg-cli <https://github.com/RockX-SG/rockx-dkg-cli>
 *CAUTION: THESE CODES HAVE NOT BEEN AUDITED
 *
 *rockx-dkg-cli is free software: you can redistribute it and/or modify
 *it under the terms of the GNU General Public License as published by
 *the Free Software Foundation, either version 3 of the License, or
 *(at your option) any later version.
 *
 *rockx-dkg-cli is distributed in the hope that it will be useful,
 *but WITHOUT ANY WARRANTY; without even the implied warranty of
 *MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *GNU General Public License for more details.
 *
 *You should have received a copy of the GNU General Public License
 *along with rockx-dkg-cli. If not, see <http://www.gnu.org/licenses/>.
 *==================================================================
 */

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
