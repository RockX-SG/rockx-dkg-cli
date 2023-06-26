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

package keymanager

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"errors"

	"github.com/bloxapp/ssv-spec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type keyManager struct {
	Domain types.DomainType
	sk     *ecdsa.PrivateKey
}

func NewKeyManager(domain types.DomainType, privateKey *ecdsa.PrivateKey) types.DKGSigner {
	return &keyManager{
		Domain: domain,
		sk:     privateKey,
	}
}

func (km *keyManager) Decrypt(sk *rsa.PrivateKey, cipher []byte) ([]byte, error) {
	if sk == nil {
		return nil, errors.New("private key is nil")
	}
	if err := sk.Validate(); err != nil {
		return nil, err
	}

	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, sk, cipher)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func (km *keyManager) Encrypt(pk *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	if pk == nil {
		return nil, errors.New("public key is nil")
	}

	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, pk, plaintext)
	if err != nil {
		return nil, err
	}
	return cipher, nil
}

func (km *keyManager) SignDKGOutput(output types.Root, address common.Address) (types.Signature, error) {
	root, err := types.ComputeSigningRoot(output, types.ComputeSignatureDomain(km.Domain, types.DKGSignatureType))
	if err != nil {
		return nil, err
	}
	return crypto.Sign(root, km.sk)
}

func (km *keyManager) SignRoot(data types.Root, sigType types.SignatureType, pk []byte) (types.Signature, error) {
	panic("not implemented")
}

func (km *keyManager) SignETHDepositRoot(root []byte, address common.Address) (types.Signature, error) {
	panic("not implemented")
}
