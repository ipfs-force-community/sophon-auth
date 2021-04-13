package storage

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/xerrors"
	"time"
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
	val, err := kp.CreateTimeBytes()
	if err != nil {
		return xerrors.Errorf("failed to marshal time :%s", err)
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(kp.Token.Bytes(), val)
	})
}

func (s *badgerStore) Delete(token Token) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(token.Bytes())
	})
}

func (s *badgerStore) Has(token Token) (bool, error) {
	var value []byte
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(token.Bytes())
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

func (s *badgerStore) List(skip, limit int64) ([]*KeyPair, error) {
	data := make(chan *KeyPair, limit)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		idx := int64(0)
		for it.Rewind(); it.Valid() && idx < skip+limit; it.Next() {
			if idx >= skip {
				item := it.Item()
				k := item.Key()
				val := new([]byte)
				err := item.Value(func(v []byte) error {
					fmt.Printf("key=%s, value=%s\n", k, v)
					val = &v
					return nil
				})
				if err != nil {
					close(data)
					return err
				}
				kp := new(KeyPair)
				err = kp.FromBytes(k, *val)
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
