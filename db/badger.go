package db

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/xerrors"
	"time"
)

type storage struct {
	db *badger.DB
}

func Open(filePath string) (Database, error) {
	db, err := badger.Open(badger.DefaultOptions(filePath))
	if err != nil {
		return nil, xerrors.Errorf("open db failed :%s", err)
	}
	s := &storage{
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

func (s *storage) SaveAndDelete(saveKey, deleteKey []byte, value []byte) (err error) {
	txn := s.db.NewTransaction(true)
	if err = txn.Set(saveKey, value); err != nil {
		txn.Discard()
		return
	}
	if err = txn.Delete(deleteKey); err != nil {
		txn.Discard()
		return
	}

	return txn.Commit()
}

func (s *storage) Put(key, value []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

//nolint
func (s *storage) BatchPut(pairs []*Pair) error {
	return s.db.Update(func(txn *badger.Txn) error {
		txn = s.db.NewTransaction(true)
		for _, pair := range pairs {
			if err := txn.Set(pair.Key, pair.Val); err == badger.ErrTxnTooBig {
				if err = txn.Commit(); err != nil {
					return err
				}
				txn = s.db.NewTransaction(true)
				if err = txn.Set(pair.Key, pair.Val); err != nil {
					return err
				}
			}
		}
		return txn.Commit()
	})
}

func (s *storage) Remove(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (s *storage) Get(key []byte) (value []byte, err error) {
	err = s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(value)
		return err
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *storage) Fetch(skip, size int64) (<-chan *Pair, error) {
	data := make(chan *Pair, size)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		idx := int64(0)
		for it.Rewind(); it.Valid() && idx < skip+size; it.Next() {
			if idx > skip {
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
				data <- &Pair{k, *val}
			}
			idx++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *storage) Iterator(prefix []byte) (<-chan *Pair, error) {
	data := make(chan *Pair, 127)
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
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
			data <- &Pair{k, *val}
		}
		close(data)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
