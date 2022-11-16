package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/log"
	"golang.org/x/xerrors"
)

type Prefix = string

const (
	PrefixToken    Prefix = "TOKEN:"
	PrefixUser     Prefix = "USER:"
	PrefixReqLimit Prefix = "ReqLimit:"
	PrefixMiner    Prefix = "MINERS:"
	PrefixSigner   Prefix = "SIGNERS:"
)

var storeVersionKey = []byte("StoreVersion")

func rateLimitKey(name string) []byte {
	return []byte(PrefixReqLimit + name)
}

func userKey(name string) []byte {
	return []byte(PrefixUser + name)
}

func tokenKey(name string) []byte {
	return []byte(PrefixToken + name)
}

func minerKey(miner string) []byte {
	return []byte(PrefixMiner + miner)
}

func signerKey(signer string) []byte {
	return []byte(PrefixSigner + signer)
}

func signerForUserKey(signer, userName string) []byte {
	return []byte(fmt.Sprintf("%s%s:%s", PrefixSigner, signer, userName))
}

// if key not exists, will get a badger.ErrKeyNotFound error.
func (s *badgerStore) delObj(key []byte) error {
	return s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return err
		}
		return txn.Delete(key)
	})
}

func (s *badgerStore) softDelObj(obj softDelete) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := obj.key()
		val, err := txn.Get(key)
		if err != nil {
			return err
		}
		if err := val.Value(func(val []byte) error {
			return obj.FromBytes(val)
		}); err != nil {
			return err
		}
		if obj.isDeleted() {
			return xerrors.Errorf("not exist")
		}
		obj.setDeleted()
		data, err := obj.Bytes()
		if err != nil {
			return xerrors.Errorf("failed to marshal time :%s", err)
		}
		return txn.Set(key, data)
	})
}

func (s *badgerStore) isExist(obj deleteVerify) (bool, error) {
	var exist bool
	err := s.db.View(func(txn *badger.Txn) error {
		key := obj.key()
		val, err := txn.Get(key)
		if err != nil {
			if xerrors.Is(err, badger.ErrKeyNotFound) {
				exist = false
				return nil
			}
			return xerrors.Errorf("get key failed:%v", err)
		}
		if err := val.Value(func(val []byte) error {
			return obj.FromBytes(val)
		}); err != nil {
			return err
		}

		exist = !obj.isDeleted()
		return nil
	})
	return exist, err
}

func (s *badgerStore) putBadgerObj(obj iBadgerObj) error {
	key := obj.key()
	return s.put(key, obj)
}

func (s *badgerStore) put(key []byte, val iStreamableObj) error {
	data, err := val.Bytes()
	if err != nil {
		return xerrors.Errorf("failed to marshal time :%s", err)
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, data)
	})
}

func (s *badgerStore) getUsableObj(key []byte, obj deleteVerify) error {
	return s.db.View(func(txn *badger.Txn) error {
		val, err := txn.Get(key)
		if err != nil {
			return err
		}
		if err := val.Value(func(val []byte) error {
			return obj.FromBytes(val)
		}); err != nil {
			return err
		}
		if obj.isDeleted() {
			return badger.ErrKeyNotFound
		}
		return nil
	})
}

func (s *badgerStore) getObj(key []byte, obj iStreamableObj) error {
	return s.db.View(func(txn *badger.Txn) error {
		val, err := txn.Get(key)
		if err != nil {
			return err
		}
		return val.Value(func(val []byte) error {
			return obj.FromBytes(val)
		})
	})
}

type fWalkCallback func(item *badger.Item) (isContinueWalk bool, err error)

func (s *badgerStore) walkThroughPrefix(prefix []byte, callback fWalkCallback) error {
	return s.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			isContinue, err := callback(it.Item())
			if err != nil {
				return err
			}
			if !isContinue {
				break
			}
		}
		return nil
	})
}

func (s *badgerStore) Version() (uint64, error) {
	var version StoreVersion
	err := s.getObj(storeVersionKey, &version)
	if err != nil {
		if xerrors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return version.Version, nil
}

func (s *badgerStore) MigrateToV1() error {
	var users []*struct {
		Miner, Name string
	}

	prefix := []byte(PrefixUser)
	now := time.Now()

	return s.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := it.Item().Value(func(val []byte) error {
				var u struct{ Miner, Name string }
				if err := json.Unmarshal(val, &u); err != nil {
					return err
				}
				users = append(users, &u)
				return nil
			}); err != nil {
				return err
			}
		}

		for _, u := range users {
			maddr, err := address.NewFromString(u.Miner)
			if err != nil || maddr.Empty() {
				log.Warnf("won't migrate miner:%s, invalid miner address", u.Miner)
				continue
			}
			b, err := (&Miner{
				Miner:        storedAddress(maddr),
				User:         u.Name,
				OrmTimestamp: OrmTimestamp{CreatedAt: now, UpdatedAt: now},
			}).Bytes()
			if err != nil {
				return err
			}
			if err := txn.Set(minerKey(u.Miner), b); err != nil {
				return err
			}
		}

		version, err := (&StoreVersion{ID: 1, Version: 1}).Bytes()
		if err != nil {
			return err
		}
		return txn.Set(storeVersionKey, version)
	})
}

func (s *badgerStore) MigrateToV2() error {
	return s.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek([]byte(PrefixMiner)); it.ValidForPrefix([]byte(PrefixMiner)); it.Next() {
			item := it.Item()
			var m Miner
			if err := item.Value(func(val []byte) error {
				if err := m.FromBytes(val); err != nil {
					return err
				}

				openMining := true
				m.OpenMining = &openMining
				b, err := m.Bytes()
				if err != nil {
					return xerrors.Errorf("get miner object data failed:%w", err)
				}
				if err := txn.Set(minerKey(m.Miner.Address().String()), b); err != nil {
					return err
				}

				return nil
			}); err != nil {
				return err
			}
		}

		version, err := (&StoreVersion{ID: 1, Version: 2}).Bytes()
		if err != nil {
			return err
		}
		return txn.Set(storeVersionKey, version)
	})
}

func (s *badgerStore) MigrateToV3() error {
	return s.db.Update(func(txn *badger.Txn) error {
		version, err := (&StoreVersion{ID: 1, Version: 3}).Bytes()
		if err != nil {
			return err
		}
		return txn.Set(storeVersionKey, version)
	})
}
