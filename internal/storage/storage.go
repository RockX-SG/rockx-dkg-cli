package storage

import (
	"encoding/json"
	"fmt"

	"github.com/bloxapp/ssv-spec/dkg"
	"github.com/bloxapp/ssv-spec/types"
	"github.com/dgraph-io/badger/v3"
)

var (
	Network = "prater"
)

type Storage struct {
	db *badger.DB
}

func NewStorage(db *badger.DB) dkg.Storage {
	return &Storage{
		db: db,
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
	return true, operator, nil
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
