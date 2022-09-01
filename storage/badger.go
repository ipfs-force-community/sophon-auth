package storage

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
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
	s := &badgerStore{db: db}
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
	return s.putBadgerObj(kp)
}

func (s *badgerStore) Delete(token Token) error {
	kp := &KeyPair{Token: token}
	return s.softDelObj(kp)
}

func (s *badgerStore) Has(token Token) (bool, error) {
	kp := &KeyPair{Token: token}
	return s.isExist(kp)
}

func (s *badgerStore) Get(token Token) (*KeyPair, error) {
	var kp KeyPair
	return &kp, s.getUsableObj(tokenKey(token.String()), &kp)
}

func (s *badgerStore) GetTokenRecord(token Token) (*KeyPair, error) {
	var kp KeyPair
	return &kp, s.getObj(tokenKey(token.String()), &kp)
}

func (s *badgerStore) ByName(name string) ([]*KeyPair, error) {
	var kps []*KeyPair
	if err := s.walkThroughPrefix([]byte(PrefixToken), func(item *badger.Item) (bool, error) {
		if err := item.Value(func(val []byte) error {
			kp := new(KeyPair)
			if err := kp.FromBytes(val); err != nil {
				return err
			}
			if kp.Name == name && !kp.isDeleted() {
				kps = append(kps, kp)
			}
			return nil
		}); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	return kps, nil
}

func (s *badgerStore) UpdateToken(kp *KeyPair) error {
	return s.putBadgerObj(kp)
}

func (s *badgerStore) List(skip, limit int64) ([]*KeyPair, error) {
	var offset int64
	var kps []*KeyPair
	if err := s.walkThroughPrefix([]byte(PrefixToken), func(item *badger.Item) (bool, error) {
		if err := item.Value(func(val []byte) error {
			kp := new(KeyPair)
			if err := kp.FromBytes(val); err != nil {
				return err
			}
			if kp.isDeleted() {
				return nil
			}
			offset++
			if offset <= skip {
				return nil
			}
			kps = append(kps, kp)
			return nil
		}); err != nil {
			return false, err
		}

		return limit == 0 || offset-skip < limit, nil
	}); err != nil {
		return nil, err
	}
	return kps, nil
}

func (s *badgerStore) GetUser(name string) (*User, error) {
	user := new(User)
	return user, s.getUsableObj(userKey(name), user)
}

func (s *badgerStore) GetUserRecord(name string) (*User, error) {
	user := new(User)
	return user, s.getObj(userKey(name), user)
}

func (s *badgerStore) UpdateUser(user *User) error {
	return s.putBadgerObj(user)
}

func (s *badgerStore) HasUser(name string) (bool, error) {
	user := &User{Name: name}
	return s.isExist(user)
}

func (s *badgerStore) PutUser(user *User) error {
	return s.putBadgerObj(user)
}

func (s *badgerStore) ListUsers(skip, limit int64, state core.UserState) ([]*User, error) {
	var users []*User
	var satisfiedItemCount = int64(0)
	if err := s.walkThroughPrefix([]byte(PrefixUser), func(item *badger.Item) (bool, error) {
		err := item.Value(func(val []byte) error {
			var user = new(User)
			if err := user.FromBytes(val); err != nil {
				return err
			}
			if state != core.UserStateUndefined && user.State != state {
				return nil
			}
			if user.isDeleted() {
				return nil
			}
			satisfiedItemCount++
			if satisfiedItemCount <= skip {
				return nil
			}
			users = append(users, user)
			return nil
		})
		return limit == 0 || satisfiedItemCount-skip < limit, err
	}); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *badgerStore) DeleteUser(name string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		user := &User{}
		key := userKey(name)
		val, err := txn.Get(key)
		if err != nil {
			return err
		}
		if err := val.Value(func(val []byte) error {
			return user.FromBytes(val)
		}); err != nil {
			return err
		}
		if user.isDeleted() {
			return xerrors.Errorf("user not exist")
		}
		user.setDeleted()
		data, err := user.Bytes()
		if err != nil {
			return err
		}
		if err := txn.Set(key, data); err != nil {
			return err
		}

		// delete miners
		miners, err := s.ListMiners(name)
		if err != nil {
			return err
		}
		addrs := make([]address.Address, 0, len(miners))
		for _, miner := range miners {
			mKey := minerKey(miner.Miner.Address().String())
			val, err := txn.Get(mKey)
			if err != nil {
				return err
			}
			m := &Miner{}
			if err := val.Value(func(val []byte) error {
				return m.FromBytes(val)
			}); err != nil {
				return err
			}
			m.setDeleted()
			data, err := m.Bytes()
			if err != nil {
				return err
			}
			if err := txn.Set(mKey, data); err != nil {
				return err
			}
			addrs = append(addrs, miner.Miner.Address())
		}
		log.Infof("delete user %s, delete miners %v", name, addrs)

		return nil
	})
}

func (s *badgerStore) GetRateLimits(name, id string) ([]*UserRateLimit, error) {
	mRateLimits, err := s.listRateLimits(name, id)
	if err != nil {
		if xerrors.Is(err, badger.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var rateLimits = make([]*UserRateLimit, len(mRateLimits))
	var idx = 0
	for _, l := range mRateLimits {
		rateLimits[idx] = l
		idx++
	}

	return rateLimits, err
}

func (s *badgerStore) PutRateLimit(limit *UserRateLimit) (string, error) {
	if limit.Id == "" {
		limit.Id = uuid.NewString()
	}
	limits, err := s.listRateLimits(limit.Name, "")
	if err != nil {
		if !xerrors.Is(err, badger.ErrKeyNotFound) {
			return "", err
		}
		limits = make(map[string]*UserRateLimit)
	}

	limits[limit.Id] = limit

	return limit.Id, s.updateUserRateLimit(limit.Name, limits)
}

func (s *badgerStore) DelRateLimit(name, id string) error {
	if len(name) == 0 || len(id) == 0 {
		return errors.New("user and rate-limit id is required for removing rate limit regulation")
	}
	mRateLimit, err := s.listRateLimits(name, id)
	if err != nil {
		return err
	}
	if _, exist := mRateLimit[id]; !exist {
		return nil
	}

	delete(mRateLimit, id)

	if len(mRateLimit) == 0 {
		return s.delObj(rateLimitKey(name))
	}
	return s.updateUserRateLimit(name, mRateLimit)
}

func (s *badgerStore) listRateLimits(user, id string) (map[string]*UserRateLimit, error) {
	var mRateLimits mapedRatelimit
	if err := s.getObj(rateLimitKey(user), &mRateLimits); err != nil {
		return nil, err
	}

	if len(id) != 0 {
		res := make(map[string]*UserRateLimit)
		if rl, exists := mRateLimits[id]; exists {
			res[id] = rl
		}
		mRateLimits = res
	}
	return mRateLimits, nil
}

func (s *badgerStore) updateUserRateLimit(name string, limits mapedRatelimit) error {
	return s.put(rateLimitKey(name), &limits)
}

// miner
func (s *badgerStore) getMiner(maddr address.Address) (*Miner, error) {
	var miner Miner
	if err := s.getUsableObj(minerKey(maddr.String()), &miner); err != nil {
		return nil, err
	}
	return &miner, nil
}

func (s *badgerStore) GetUserByMiner(mAddr address.Address) (*User, error) {
	miner, err := s.getMiner(mAddr)
	if err != nil {
		return nil, err
	}
	return s.GetUser(miner.User)
}

func (s *badgerStore) UpsertMiner(maddr address.Address, userName string, openMining bool) (bool, error) {
	miner := &Miner{}
	now := time.Now()
	var isCreate bool
	userkey, minerkey := userKey(userName), minerKey(maddr.String())
	return isCreate, s.db.Update(func(txn *badger.Txn) error {
		// this 'get(userKey)' purpose to makesure 'user' exist
		if _, err := txn.Get(userkey); err != nil {
			if xerrors.Is(err, badger.ErrKeyNotFound) {
				return xerrors.Errorf("can't bind miner:%s to not exist user:%s",
					maddr.String(), userName)
			}
			return xerrors.Errorf("bound miner:%s to user:%s failed, %w",
				maddr.String(), userName, err)
		}

		// if miner already exists, update it
		if item, err := txn.Get(minerkey); err != nil {
			if xerrors.Is(err, badger.ErrKeyNotFound) {
				miner.Miner = storedAddress(maddr)
				miner.CreatedAt = now
				isCreate = true
			} else {
				return err
			}
		} else {
			if err = item.Value(func(val []byte) error { return miner.FromBytes(val) }); err != nil {
				return err
			}
		}
		miner.User = userName
		miner.OpenMining = &openMining
		miner.UpdatedAt = now
		// update miner to valid
		miner.DeletedAt.Valid = true
		miner.DeletedAt.Time = time.Time{}

		val, err := miner.Bytes()
		if err != nil {
			return xerrors.Errorf("get miner object data failed:%w", err)
		}
		return txn.Set(minerkey, val)
	})
}

func (s *badgerStore) HasMiner(maddr address.Address) (bool, error) {
	miner := &Miner{Miner: storedAddress(maddr)}
	return s.isExist(miner)
}

func (s *badgerStore) MinerExistInUser(maddr address.Address, userName string) (bool, error) {
	bHas := false
	if err := s.walkThroughPrefix([]byte(PrefixMiner), func(item *badger.Item) (isContinueWalk bool, err error) {
		var miner Miner
		if err := item.Value(func(val []byte) error {
			if err := miner.FromBytes(val); err != nil {
				return err
			}

			if miner.User == userName && miner.Miner.Address().String() == maddr.String() && !miner.isDeleted() {
				bHas = true
			}

			return nil
		}); err != nil {
			return false, err
		}
		return !bHas, nil
	}); err != nil {
		return false, err
	}

	return bHas, nil
}

func (s *badgerStore) GetUserBySigner(addr address.Address) (*User, error) {
	var signer Signer
	if err := s.getUsableObj(signerKey(addr.String()), &signer); err != nil {
		return nil, err
	}

	return s.GetUser(signer.User)
}

func (s *badgerStore) ListMiners(user string) ([]*Miner, error) {
	var miners []*Miner
	if err := s.walkThroughPrefix([]byte(PrefixMiner), func(item *badger.Item) (isContinueWalk bool, err error) {
		var m Miner
		if err := item.Value(func(val []byte) error {
			if err := m.FromBytes(val); err != nil {
				return err
			}
			if m.User == user && !m.isDeleted() {
				miners = append(miners, &m)
			}
			return nil
		}); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		return nil, err
	}
	return miners, nil
}

func (s *badgerStore) DelMiner(miner address.Address) (bool, error) {
	m := &Miner{Miner: storedAddress(miner)}
	err := s.softDelObj(m)
	if err != nil {
		if xerrors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *badgerStore) UpsertSigner(addr address.Address, userName string) (bool, error) {
	signer := &Signer{}
	now := time.Now()
	var isCreate bool
	userKey, signerKey := userKey(userName), signerKey(addr.String())
	return isCreate, s.db.Update(func(txn *badger.Txn) error {
		// this 'get(userKey)' purpose to make sure 'user' exist
		if _, err := txn.Get(userKey); err != nil {
			if xerrors.Is(err, badger.ErrKeyNotFound) {
				return xerrors.Errorf("can't bind signer:%s to not exist user:%s",
					addr.String(), userName)
			}
			return xerrors.Errorf("bound signer:%s to user:%s failed, %w",
				addr.String(), userName, err)
		}

		// if signer already exists, update it
		if item, err := txn.Get(signerKey); err != nil {
			if xerrors.Is(err, badger.ErrKeyNotFound) {
				signer.Signer = storedAddress(addr)
				signer.CreatedAt = now
				isCreate = true
			} else {
				return err
			}
		} else {
			if err = item.Value(func(val []byte) error { return signer.FromBytes(val) }); err != nil {
				return err
			}
		}
		signer.User = userName
		signer.UpdatedAt = now
		signer.DeletedAt.Valid = true
		signer.DeletedAt.Time = time.Time{}

		val, err := signer.Bytes()
		if err != nil {
			return xerrors.Errorf("get signer object data failed:%w", err)
		}
		return txn.Set(signerKey, val)
	})
}

func (s *badgerStore) HasSigner(addr address.Address, userName string) (bool, error) {
	bHas := false
	if err := s.walkThroughPrefix([]byte(PrefixSigner), func(item *badger.Item) (isContinueWalk bool, err error) {
		var signer Signer
		if err := item.Value(func(val []byte) error {
			if err := signer.FromBytes(val); err != nil {
				return err
			}

			if len(userName) > 0 {
				if signer.User == userName && signer.Signer.Address().String() == addr.String() && !signer.isDeleted() {
					bHas = true
				}
			} else {
				if signer.Signer.Address().String() == addr.String() && !signer.isDeleted() {
					bHas = true
				}
			}
			return nil
		}); err != nil {
			return false, err
		}
		return !bHas, nil
	}); err != nil {
		return false, err
	}

	return bHas, nil
}

func (s *badgerStore) ListSigner(user string) ([]*Signer, error) {
	var signers []*Signer
	if err := s.walkThroughPrefix([]byte(PrefixSigner), func(item *badger.Item) (isContinueWalk bool, err error) {
		var signer Signer
		if err := item.Value(func(val []byte) error {
			if err := signer.FromBytes(val); err != nil {
				return err
			}
			if signer.User == user && !signer.isDeleted() {
				signers = append(signers, &signer)
			}
			return nil
		}); err != nil {
			return false, err
		}
		return true, nil
	}); err != nil {
		return nil, err
	}
	return signers, nil
}

func (s *badgerStore) DelSigner(addr address.Address) (bool, error) {
	signer := &Signer{Signer: storedAddress(addr)}
	err := s.softDelObj(signer)
	if err != nil {
		if xerrors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
