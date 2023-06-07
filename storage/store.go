package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/log"
)

func NewStore(cnf *config.DBConfig, dataPath string) (Store, error) {
	var store Store
	var err error
	switch strings.ToLower(cnf.Type) {
	case config.Mysql:
		log.Warn("mysql storage")
		store, err = newMySQLStore(cnf)
	case config.Badger:
		log.Warn("badger storage")
		store, err = newBadgerStore(dataPath)
	default:
		return nil, fmt.Errorf("the type %s is not currently supported", cnf.Type)
	}

	if err != nil {
		return nil, err
	}
	if err = StoreMigrate(store); err != nil {
		return nil, xerrors.Errorf("migrate store failed:%w", err)
	}

	return store, nil
}

type Store interface {
	// token
	Get(token Token) (*KeyPair, error)
	ByName(name string) ([]*KeyPair, error)
	Put(kp *KeyPair) error
	Delete(token Token) error
	Recover(token Token) error
	Has(token Token) (bool, error)
	List(skip, limit int64) ([]*KeyPair, error)
	UpdateToken(kp *KeyPair) error

	// user
	HasUser(name string) (bool, error)
	GetUser(name string) (*User, error)
	VerifyUsers(names []string) error
	PutUser(*User) error
	UpdateUser(*User) error
	ListUsers(skip, limit int64, state core.UserState) ([]*User, error)
	DeleteUser(name string) error
	RecoverUser(name string) error

	// rate limit
	GetRateLimits(name, id string) ([]*UserRateLimit, error)
	PutRateLimit(limit *UserRateLimit) (string, error)
	DelRateLimit(name, id string) error

	// miner-user(1-1)
	// first returned bool, 'miner' is created(true) or updated(false)
	UpsertMiner(mAddr address.Address, userName string, openMining *bool) (bool, error)
	HasMiner(mAddr address.Address) (bool, error)
	MinerExistInUser(mAddr address.Address, userName string) (bool, error)
	GetUserByMiner(mAddr address.Address) (*User, error)
	ListMiners(user string) ([]*Miner, error)
	// first returned bool, if miner exists(true) or false
	DelMiner(mAddr address.Address) (bool, error)

	// signer-user(n-n)
	RegisterSigner(addr address.Address, userName string) error
	SignerExistInUser(addr address.Address, userName string) (bool, error)
	ListSigner(userName string) ([]*Signer, error)
	UnregisterSigner(addr address.Address, userName string) error
	// has signer in system
	HasSigner(addr address.Address) (bool, error)
	// delete all signers
	DelSigner(addr address.Address) (bool, error)
	// all users including the specified signer
	GetUserBySigner(addr address.Address) ([]*User, error)

	Version() (uint64, error)
	MigrateToV1() error
	MigrateToV2() error
	MigrateToV3() error
}

type KeyPair struct {
	Name       string    `gorm:"column:name;type:varchar(50);NOT NULL"`
	Perm       string    `gorm:"column:perm;type:varchar(50);NOT NULL"`
	Secret     string    `gorm:"column:secret;type:varchar(255);NOT NULL"`
	Extra      string    `gorm:"column:extra;type:varchar(255);"`
	Token      Token     `gorm:"column:token;type:varchar(512);uniqueIndex:token_token_IDX,type:hash;not null"`
	CreateTime time.Time `gorm:"column:createTime;type:datetime;NOT NULL"`
	IsDeleted  int       `gorm:"column:is_deleted;index;default:0;NOT NULL"`
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

func (kp *KeyPair) Bytes() ([]byte, error) {
	buff, err := json.Marshal(kp)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (kp *KeyPair) CreateTimeBytes() ([]byte, error) {
	val, err := kp.CreateTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (kp *KeyPair) FromBytes(val []byte) error {
	return json.Unmarshal(val, kp)
}

func (kp *KeyPair) key() []byte {
	return tokenKey(kp.Token.String())
}

func (kp *KeyPair) isDeleted() bool {
	return kp.IsDeleted == core.Deleted
}

func (kp *KeyPair) setDeleted() {
	kp.IsDeleted = core.Deleted
}

type User struct {
	Id         string         `gorm:"column:id;type:varchar(64);primary_key"`
	Name       string         `gorm:"column:name;type:varchar(50);uniqueIndex:users_name_IDX,type:btree;not null"`
	Comment    string         `gorm:"column:comment;type:varchar(255);"`
	State      core.UserState `gorm:"column:state;type:tinyint(4);default:0;NOT NULL"`
	CreateTime time.Time      `gorm:"column:createTime;type:datetime;NOT NULL"`
	UpdateTime time.Time      `gorm:"column:updateTime;type:datetime;NOT NULL"`
	IsDeleted  int            `gorm:"column:is_deleted;index;default:0;NOT NULL"`
}

type OrmTimestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	// Implemented soft delete
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (*User) TableName() string {
	return "users"
}

func (u *User) key() []byte {
	return userKey(u.Name)
}

func (u *User) Bytes() ([]byte, error) {
	buff, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func (u *User) FromBytes(buff []byte) error {
	return json.Unmarshal(buff, u)
}

func (u *User) CreateTimeBytes() ([]byte, error) {
	val, err := u.CreateTime.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (u *User) isDeleted() bool {
	return u.IsDeleted == 1
}

func (u *User) setDeleted() {
	u.IsDeleted = core.Deleted
}

type storedAddress address.Address

func (sa storedAddress) Address() address.Address {
	return address.Address(sa)
}

func (sa storedAddress) String() string {
	val := sa.Address().String()
	if sa.Address().Empty() {
		return val
	}
	return val[1:]
}

func (sa *storedAddress) Scan(value interface{}) error {
	val, isok := value.([]byte)
	if !isok {
		return xerrors.New("non-string types unsupported")
	}
	var str string
	if address.CurrentNetwork == address.Mainnet {
		str = address.MainnetPrefix + string(val)
	} else {
		str = address.TestnetPrefix + string(val)
	}

	addr, err := address.NewFromString(str)
	if err != nil {
		return err
	}
	*sa = storedAddress(addr)
	return nil
}

func (sa *storedAddress) UnmarshalJSON(b []byte) error {
	return (*address.Address)(sa).UnmarshalJSON(b)
}

// MarshalJSON implements the json marshal interface.
func (sa storedAddress) MarshalJSON() ([]byte, error) {
	return address.Address(sa).MarshalJSON()
}

func (sa storedAddress) Value() (driver.Value, error) {
	val := sa.String()
	if sa.Address().Empty() {
		return val, nil
	}
	return sa.String(), nil
}

type Miner struct {
	ID         uint64        `gorm:"column:id;primary_key;bigint(20) unsigned AUTO_INCREMENT"`
	Miner      storedAddress `gorm:"column:miner;type:varchar(128);uniqueIndex:miner_idx;NOT NULL"`
	User       string        `gorm:"column:user;type:varchar(50);NOT NULL"`
	OpenMining *bool         `gorm:"column:open_mining;default:1;comment:0-false,1-true"`
	OrmTimestamp
}

func (m *Miner) Bytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Miner) FromBytes(buf []byte) error {
	return json.Unmarshal(buf, m)
}

func (m *Miner) key() []byte {
	return minerKey(m.Miner.Address().String())
}

func (m *Miner) isDeleted() bool {
	return m.DeletedAt.Valid && !m.DeletedAt.Time.IsZero()
}

func (m *Miner) setDeleted() {
	m.DeletedAt.Valid = true
	m.DeletedAt.Time = time.Now()
}

type Signer struct {
	ID     uint64        `gorm:"column:id;primary_key;bigint(20) unsigned AUTO_INCREMENT;"`
	Signer storedAddress `gorm:"column:signer;type:varchar(128);uniqueIndex:user_signer_idx,priority:2;NOT NULL"`
	User   string        `gorm:"column:user;type:varchar(50);uniqueIndex:user_signer_idx,priority:1;NOT NULL"`
	OrmTimestamp
}

func (m *Signer) Bytes() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Signer) FromBytes(buf []byte) error {
	return json.Unmarshal(buf, m)
}

func (m *Signer) key() []byte {
	return signerForUserKey(m.Signer.Address().String(), m.User)
}

func (m *Signer) isDeleted() bool {
	return m.DeletedAt.Valid && !m.DeletedAt.Time.IsZero()
}

func (m *Signer) setDeleted() {
	m.DeletedAt.Valid = true
	m.DeletedAt.Time = time.Now()
}

type StoreVersion struct {
	ID      uint64 `grom:"primary_key"`
	Version uint64 `gorm:"column:version"`
}

func (s *StoreVersion) key() []byte {
	return storeVersionKey
}

func (s *StoreVersion) FromBytes(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (s *StoreVersion) Bytes() ([]byte, error) {
	return json.Marshal(s)
}

// we are perpose to support limit user requests with `Service`/`Service.API` ferther,
// so we add their declar in `UserRateLimit`
type UserRateLimit struct {
	Id       string   `gorm:"column:id;type:varchar(64);primary_key"`
	Name     string   `gorm:"column:name;type:varchar(50);index:user_service_api_IDX;not null" binding:"required"`
	Service  string   `gorm:"column:service;type:varchar(50);index:user_service_api_IDX"`
	API      string   `gorm:"column:api;type:varchar(50);index:user_service_api_IDX"`
	ReqLimit ReqLimit `gorm:"column:reqLimit;type:varchar(256)"`
}

func (l *UserRateLimit) LimitKey() string {
	return l.Name + l.Service + l.API
}

type ReqLimit struct {
	Cap      int64
	ResetDur time.Duration
}

func (rl *ReqLimit) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return xerrors.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	if len(bytes) == 0 {
		*rl = ReqLimit{}
		return nil
	}
	return json.Unmarshal(bytes, rl)
}

func (rl ReqLimit) Value() (driver.Value, error) {
	return json.Marshal(rl)
}

type mapedRatelimit map[string]*UserRateLimit

// todo: should think about if `mapedRatelimte` is empty?
func (mr *mapedRatelimit) key() []byte {
	for _, v := range *mr {
		return rateLimitKey(v.Name)
	}
	return nil
}

func (mr *mapedRatelimit) FromBytes(buf []byte) error {
	return json.Unmarshal(buf, mr)
}

func (mr *mapedRatelimit) Bytes() ([]byte, error) {
	return json.Marshal(mr)
}

type iKableObj interface {
	key() []byte
}

type iStreamableObj interface {
	FromBytes([]byte) error
	Bytes() ([]byte, error)
}

type deleteVerify interface {
	iKableObj
	iStreamableObj
	// isDeleted return whether data is deleted, true if deleted
	isDeleted() bool
}

type softDelete interface {
	deleteVerify
	// SetDeleted set data to deleted
	setDeleted()
}

type iBadgerObj interface {
	iKableObj
	iStreamableObj
}

var (
	_ iBadgerObj = (*Miner)(nil)
	_ iBadgerObj = (*User)(nil)
	_ iBadgerObj = (*KeyPair)(nil)
	_ iBadgerObj = (*mapedRatelimit)(nil)
	_ iBadgerObj = (*StoreVersion)(nil)
)
