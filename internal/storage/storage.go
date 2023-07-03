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

package storage

import (
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/dgraph-io/badger/v3"
	"github.com/herumi/bls-eth-go-binary/bls"
)

var (
	Network = "prater"
)

type Storage struct {
	db           *badger.DB
	thisOperator types.OperatorID
	thisSK       *rsa.PrivateKey
}

func NewStorage(db *badger.DB, operatorID types.OperatorID, operatorKey *rsa.PrivateKey) dkg.Storage {
	return &Storage{
		db:           db,
		thisOperator: operatorID,
		thisSK:       operatorKey,
	}
}

func (s *Storage) GetDKGOperator(operatorID types.OperatorID) (bool, *dkg.Operator, error) {

	var (
		val          []byte
		requireFetch bool   = false
		key          string = fmt.Sprintf("operator/%d", operatorID)
	)

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(nil)
		return err
	})
	if err == badger.ErrKeyNotFound {
		requireFetch = true
	} else if err != nil {
		return false, nil, err
	}

	var operator = new(dkg.Operator)
	if !requireFetch {
		if err := json.Unmarshal(val, operator); err != nil {
			return false, nil, err
		}
	} else {
		operator, err = FetchOperatorByID(operatorID)
		if err != nil {
			return false, nil, err
		}
		value, err := json.Marshal(operator)
		if err != nil {
			return false, nil, fmt.Errorf("failed to marshal keygen output :: %s", err.Error())
		}
		if err = s.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(key), value)
		}); err != nil {
			return false, nil, err
		}
	}

	if operatorID == s.thisOperator {
		operator.EncryptionPrivateKey = s.thisSK
	}

	return true, operator, nil
}

type KeyGenOutput struct {
	Share           string
	OperatorPubKeys map[types.OperatorID]string
	ValidatorPK     string
	Threshold       uint64
}

func (o *KeyGenOutput) Encode(output *dkg.KeyGenOutput) ([]byte, error) {
	kgo := &KeyGenOutput{
		Share:           output.Share.SerializeToHexStr(),
		OperatorPubKeys: make(map[types.OperatorID]string),
		ValidatorPK:     hex.EncodeToString(output.ValidatorPK),
		Threshold:       output.Threshold,
	}
	for operatorID, pk := range output.OperatorPubKeys {
		kgo.OperatorPubKeys[operatorID] = pk.SerializeToHexStr()
	}
	return json.Marshal(kgo)
}

func (o *KeyGenOutput) Decode(output []byte) (*dkg.KeyGenOutput, error) {
	if err := json.Unmarshal(output, o); err != nil {
		return nil, err
	}

	kgo := &dkg.KeyGenOutput{
		OperatorPubKeys: make(map[types.OperatorID]*bls.PublicKey),
		Threshold:       o.Threshold,
	}

	vk, err := hex.DecodeString(o.ValidatorPK)
	if err != nil {
		return nil, err
	}
	kgo.ValidatorPK = vk

	share := bls.SecretKey{}
	if err := share.DeserializeHexStr(o.Share); err != nil {
		return nil, err
	}
	kgo.Share = &share

	for operatorID, pkhex := range o.OperatorPubKeys {
		pk := bls.PublicKey{}
		if err := pk.DeserializeHexStr(pkhex); err != nil {
			return nil, err
		}
		kgo.OperatorPubKeys[operatorID] = &pk
	}
	return kgo, nil
}

func (s *Storage) SaveKeyGenOutput(output *dkg.KeyGenOutput) error {
	kgo := &KeyGenOutput{}
	value, err := kgo.Encode(output)
	if err != nil {
		return fmt.Errorf("failed to marshal keygen output :: %s", err.Error())
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(output.ValidatorPK), value)
	})
}

func (s *Storage) GetKeyGenOutput(pk types.ValidatorPK) (*dkg.KeyGenOutput, error) {
	var val []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(pk))
		if err != nil {
			return err
		}

		val, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	kgo := &KeyGenOutput{}
	result, err := kgo.Decode(val)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal keygen output :: %s", err.Error())
	}
	return result, nil
}
