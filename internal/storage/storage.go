package storage

import (
	"encoding/json"
	"fmt"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/dgraph-io/badger/v3"
)

type Storage struct {
	db        *badger.DB
	operators map[types.OperatorID]*dkg.Operator
}

func NewStorage(db *badger.DB) dkg.Storage {
	operators := make(map[types.OperatorID]*dkg.Operator)
	for operatorID, operator := range DKGOperators {
		operators[operatorID] = &dkg.Operator{
			OperatorID:       operatorID,
			ETHAddress:       operator.ETHAddress,
			EncryptionPubKey: &operator.EncryptionKey.PublicKey,
		}
	}

	return &Storage{
		db:        db,
		operators: operators,
	}
}

func (s *Storage) GetDKGOperator(operatorID types.OperatorID) (bool, *dkg.Operator, error) {
	if ret, found := s.operators[operatorID]; found {
		return true, ret, nil
	}
	return false, nil, nil
}

func (s *Storage) SaveKeyGenOutput(output *dkg.KeyGenOutput) error {
	value, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal keygen output :: %s", err.Error())
	}

	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(output.ValidatorPK, value)
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

	result := &dkg.KeyGenOutput{}
	if err = json.Unmarshal(val, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keygen output :: %s", err.Error())
	}
	return result, nil
}
