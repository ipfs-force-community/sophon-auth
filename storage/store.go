package storage

import (
	"fmt"
	"github.com/ipfs-force-community/venus-auth/config"
	"github.com/ipfs-force-community/venus-auth/log"
	"strings"
	"time"
)

func NewStore(cnf *config.DBConfig, dataPath string) (Store, error) {

	switch strings.ToLower(cnf.Type) {
	case config.Mysql:
		log.Warn("mysql storage")
		return newMySQLStore(cnf)
	case config.Badger:
		log.Warn("badger storage")
		return newBadgerStore(dataPath)
	}
	return nil, fmt.Errorf("the type %s is not currently supported", cnf.Type)
}

type Store interface {
	Put(kp *KeyPair) error
	Delete(token Token) error
	Has(token Token) (bool, error)
	List(skip, limit int64) ([]*KeyPair, error)
}

type KeyPair struct {
	Token      Token     `db:"token"`
	CreateTime time.Time `db:"createTime"`
	Perm       string    `db:"perm"`
	Name       string    `db:"name"`
	Extra      string    `db:"extra"`
}

type Token string

func (t Token) Bytes() []byte {
	return []byte(t)
}
func (t Token) String() string {
	return string(t)
}

func (t *KeyPair) CreateTimeBytes() ([]byte, error) {
	val, err := t.CreateTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (t *KeyPair) FromBytes(key []byte, val []byte) error {
	t.Token = Token(key)
	tm := time.Time{}
	err := tm.UnmarshalBinary(val)
	if err != nil {
		return err
	}
	t.CreateTime = tm
	return nil
}
