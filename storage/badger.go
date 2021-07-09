package storage

import (
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/filecoin-project/go-address"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-auth/core"
)

var _ Store = &badgerStore{}

type badgerStore struct {
	db *badger.DB
}

func newBadgerStore(filePath string) (Store, error) {
	db, err := badger.Open(badger.DefaultOptions(filePath))
	if err != nil {
		return nil, xerrors.Errorf("open db failed :%s", err)
	}
	s := &badgerStore{
		db: db,
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
		again:
			err := db.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}()
	return s, nil
}

func (s *badgerStore) Put(kp *KeyPair) error {
	val, err := kp.Bytes()
	if err != nil {
		return xerrors.Errorf("failed to marshal time :%s", err)
	}
	key := s.tokenKey(kp.Token.String())
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (s *badgerStore) Delete(token Token) error {
	key := s.tokenKey(token.String())
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (s *badgerStore) Has(token Token) (bool, error) {
	key := s.tokenKey(token.String())
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(value)
		return err
	})
	if err != nil {
		if err.Error() == "Key not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *badgerStore) Get(token Token) (*KeyPair, error) {
	kp := new(KeyPair)
	key := s.tokenKey(token.String())

	err := s.db.View(func(txn *badger.Txn) error {
		val, err := txn.Get(key)
		if err != nil || err == badger.ErrKeyNotFound {
			return xerrors.Errorf("token %s not exit", token)
		}

		return val.Value(func(val []byte) error {
			return kp.FromBytes(val)
		})
	})

	return kp, err
}

func (s *badgerStore) UpdateToken(kp *KeyPair) error {
	val, err := kp.Bytes()
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(s.tokenKey(kp.Token.String()), val)
	})
}

func (s *badgerStore) List(skip, limit int64) ([]*KeyPair, error) {
	data := make(chan *KeyPair, limit)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{
			PrefetchValues: true,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         []byte(PrefixToken),
		}
		it := txn.NewIterator(opts)
		defer it.Close()
		idx := int64(0)
		for it.Rewind(); it.Valid() && idx < skip+limit; it.Next() {
			if idx >= skip {
				item := it.Item()
				val := new([]byte)
				err := item.Value(func(v []byte) error {
					val = &v
					return nil
				})
				if err != nil {
					close(data)
					return err
				}
				kp := new(KeyPair)
				err = kp.FromBytes(*val)
				if err != nil {
					close(data)
					return err
				}
				data <- kp
			}
			idx++
		}
		close(data)
		return nil
	})
	if err != nil {
		return nil, err
	}
	res := make([]*KeyPair, 0, limit)
	for ch := range data {
		res = append(res, ch)
	}
	return res, nil
}

func (s *badgerStore) GetAccount(name string) (*Account, error) {
	account := new(Account)
	err := s.db.View(func(txn *badger.Txn) error {
		val, err := txn.Get(s.accountKey(name))
		if err != nil || err == badger.ErrKeyNotFound {
			return xerrors.Errorf("accounts %s not exit", name)
		}

		return val.Value(func(val []byte) error {
			return account.FromBytes(val)
		})
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *badgerStore) UpdateAccount(account *Account) error {
	old, err := s.GetAccount(account.Name)
	if err != nil {
		return err
	}
	account.Id = old.Id
	val, err := account.Bytes()
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(s.accountKey(account.Name), val)
	})
}

func (s *badgerStore) HasAccount(name string) (bool, error) {
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(s.accountKey(name))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(value)
		return err
	})
	if err != nil {
		if err.Error() == "Key not found" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *badgerStore) PutAccount(account *Account) error {
	val, err := account.Bytes()
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(s.accountKey(account.Name), val)
	})
}

func (s *badgerStore) ListAccounts(skip, limit int64, state int, sourceType core.SourceType, code core.KeyCode) ([]*Account, error) {
	var data []*Account
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{
			PrefetchValues: true,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         []byte(PrefixAccount),
		}
		it := txn.NewIterator(opts)
		defer it.Close()
		idx := int64(0)
		for it.Rewind(); it.Valid() && idx < skip+limit; it.Next() {
			if idx >= skip {
				item := it.Item()
				// k := item.Key()
				val := new([]byte)
				err := item.Value(func(v []byte) error {
					val = &v
					return nil
				})
				if err != nil {
					return err
				}
				account := new(Account)
				err = account.FromBytes(*val)
				if err != nil {
					return err
				}
				// aggregation multi-select
				need := false
				if code&1 == 1 {
					need = account.SourceType == sourceType
				} else {
					need = true
				}

				if !need {
					continue
				}
				if code&2 == 2 {
					need = need && account.State == state
				} else {
					need = need && true
				}
				if need {

					data = append(data, account)
					idx++
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *badgerStore) HasMiner(maddr address.Address) (bool, error) {
	var has bool
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{
			PrefetchValues: true,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         []byte(PrefixAccount),
		}
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			// k := item.Key()
			val := new([]byte)
			err := item.Value(func(v []byte) error {
				// fmt.Printf("key=%s, value=%s\n", k, v)
				val = &v
				return nil
			})
			if err != nil {
				return err
			}
			account := new(Account)
			err = account.FromBytes(*val)
			if err != nil {
				return err
			}

			if account.Miner == maddr.String() {
				has = true
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return has, nil
}

func (s *badgerStore) GetMiner(maddr address.Address) (*Account, error) {
	var data *Account
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{
			PrefetchValues: true,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         []byte(PrefixAccount),
		}
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			// k := item.Key()
			val := new([]byte)
			err := item.Value(func(v []byte) error {
				// fmt.Printf("key=%s, value=%s\n", k, v)
				val = &v
				return nil
			})
			if err != nil {
				return err
			}
			account := new(Account)
			err = account.FromBytes(*val)
			if err != nil {
				return err
			}
			if account.Miner == maddr.String() {
				data = account
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, xerrors.Errorf("miner %s not exit", maddr)
	}
	return data, nil
}
