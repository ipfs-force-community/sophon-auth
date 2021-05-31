package storage

import (
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
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

	//user
	HasUser(name string) (bool, error)
	GetUser(name string) (*User, error)
	PutUser(*User) error
	UpdateUser(*User) error
	ListUsers(skip, limit int64, state int, sourceType core.SourceType,code core.KeyCode) ([]*User, error)
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

type User struct {
	Id         string          `db:"id"`
	Name       string          `db:"name"`
	Miner      string          `db:"miner"` // miner address f01234
	Comment    string          `db:"comment"`
	SourceType core.SourceType `db:"stype"`
	State      int             `db:"state"` // 0: disable, 1: enable
	CreateTime time.Time       `db:"createTime"`
	UpdateTime time.Time       `db:"updateTime"`
}

func (t *User) Bytes() ([]byte, error) {
	buff, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (t *User) FromBytes(buff []byte) error {
	err := json.Unmarshal(buff, t)
	return err
}

func (t *User) CreateTimeBytes() ([]byte, error) {
	val, err := t.CreateTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return val, nil
}
