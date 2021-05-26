package storage

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/xerrors"
	"math"
	"time"
)

var _ Store = &badgerStore{}

type badgerStore struct {
	db         *badger.DB
	userPrefix string
}

func newBadgerStore(filePath string) (Store, error) {
	db, err := badger.Open(badger.DefaultOptions(filePath))
	if err != nil {
		return nil, xerrors.Errorf("open db failed :%s", err)
	}
	s := &badgerStore{
		db:         db,
		userPrefix: "users_",
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

func (s *badgerStore) GetUser(name string) (*User, error) {
	key := []byte(s.userPrefix + name)
	var user User
	err := s.db.View(func(txn *badger.Txn) error {
		val, err := txn.Get(key)
		if err == nil || err == badger.ErrKeyNotFound {
			return xerrors.Errorf("users %s not exit", name)
		}

		return val.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *badgerStore) UpdateUser(user *User) error {
	allUsers, err := s.ListUser(0, math.MaxInt64)
	if err != nil {
		return err
	}
	if len(user.Id) > 0 {
		//update
		for _, oldUser := range allUsers {
			if oldUser.Id == user.Id {
				continue
			}
			if oldUser.Name == user.Name {
				return xerrors.Errorf("user %s has exit", user.Name)
			}
		}
	} else {
		//create
		for _, oldUser := range allUsers {
			if oldUser.Name == user.Name {
				return xerrors.Errorf("user %s has exit", user.Name)
			}
		}
	}
	key := []byte(s.userPrefix + user.Name)
	return s.db.Update(func(txn *badger.Txn) error {
		//insert
		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}
		return txn.Set(key, userBytes)
	})
}

func (s *badgerStore) ListUser(skip, limit int64) ([]*User, error) {
	var data []*User
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.IteratorOptions{
			PrefetchValues: true,
			PrefetchSize:   100,
			Reverse:        false,
			AllVersions:    false,
			Prefix:         []byte(s.userPrefix),
		}

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
					return err
				}
				user := new(User)
				err = json.Unmarshal(*val, user)
				if err != nil {
					return err
				}
				data = append(data, user)
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
