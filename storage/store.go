package storage

import (
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/go-address"
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

	// user
	HasUser(name string) (bool, error)
	GetUser(name string) (*User, error)
	HasMiner(maddr address.Address) (bool, error)
	GetMiner(maddr address.Address) (*User, error)
	PutUser(*User) error
	UpdateUser(*User) error
	ListUsers(skip, limit int64, state int, sourceType core.SourceType, code core.KeyCode) ([]*User, error)
}

type KeyPair struct {
	Name       string    `gorm:"column:name;type:varchar(256);primary_key;NOT NULL"`
	Perm       string    `gorm:"column:perm;type:varchar(256);primary_key;NOT NULL"`
	Extra      string    `gorm:"column:extra;type:varchar(256);"`
	Token      Token     `gorm:"column:token;type:varchar(512);index:token_token_IDX,type:hash;unique;NOT NULL"`
	CreateTime time.Time `gorm:"column:createTime;type:datetime;index;NOT NULL"`
}

func (*KeyPair) TableName() string {
	return "token"
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
	Id         string          `gorm:"column:id;type:varchar(256);primary_key"`
	Name       string          `gorm:"column:name;type:varchar(256);unique;NOT NULL"`
	Miner      string          `gorm:"column:miner;type:varchar(256);index:users_miner_IDX;"`
	Comment    string          `gorm:"column:comment;type:varchar(256);"`
	SourceType core.SourceType `gorm:"column:stype;type:tinyint(4);default:0;NOT NULL"`
	State      int             `gorm:"column:state;type:tinyint(4);default:0;NOT NULL"`
	Burst 	   int 			   `gorm:"column:burst;type:int;default:0;NOT NULL"`
	Rate  	   int             `gorm:"column:rate;type:int;default:0;NOT NULL"`
	CreateTime time.Time       `gorm:"column:createTime;type:datetime;index;NOT NULL"`
	UpdateTime time.Time       `gorm:"column:updateTime;type:datetime;index;NOT NULL"`
}

func (*User) TableName() string {
	return "users"
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
