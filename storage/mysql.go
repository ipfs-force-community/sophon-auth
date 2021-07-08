package storage

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/util"
	"golang.org/x/xerrors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type mysqlStore struct {
	db  *gorm.DB
	pkg string
}

func newMySQLStore(cnf *config.DBConfig) (Store, error) {
	db, err := gorm.Open(mysql.Open(cnf.DSN))
	if err != nil {
		return nil, xerrors.Errorf("[db connection failed] Database name: %s %w", cnf.DSN, err)
	}
	if cnf.Debug {
		db = db.Debug()
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cnf.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cnf.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cnf.MaxLifeTime)
	sqlDB.SetConnMaxIdleTime(cnf.MaxIdleTime)

	store := &mysqlStore{
		db:  db,
		pkg: util.PackagePath(mysqlStore{}),
	}

	err = db.AutoMigrate(&KeyPair{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&User{})
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *mysqlStore) Put(kp *KeyPair) error {
	return s.db.Save(kp).Error
}

func (s mysqlStore) Delete(token Token) error {
	s.db.Delete(&KeyPair{}, "token=?", token)
	return nil
}

func (s mysqlStore) Has(token Token) (bool, error) {
	var count int64
	err := s.db.Table("token").Where("token=?", token).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlStore) List(skip, limit int64) ([]*KeyPair, error) {
	var tokens []*KeyPair
	err := s.db.Offset(int(skip)).Limit(int(limit)).Order("name").Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s mysqlStore) HasUser(name string) (bool, error) {
	var count int64
	err := s.db.Table("users").Where("name=?", name).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *mysqlStore) UpdateUser(user *User) error {
	user.UpdateTime = time.Now()
	err := s.db.Table("users").Save(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *mysqlStore) PutUser(user *User) error {
	user.UpdateTime = time.Now()
	user.CreateTime = time.Now()
	err := s.db.Table("users").Save(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *mysqlStore) ListUsers(skip, limit int64, state int, sourceType core.SourceType, code core.KeyCode) ([]*User, error) {
	exec := s.db.Table("users")
	if code&1 == 1 {
		exec = exec.Where("stype=?", sourceType)
	}
	if code&2 == 2 {
		exec = exec.Where("state=?", state)
	}
	arr := make([]*User, 0)
	err := exec.Order("createTime").Offset(int(skip)).Limit(int(limit)).Scan(&arr).Error
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (s *mysqlStore) GetUser(name string) (*User, error) {
	var user User
	err := s.db.Table("users").Take(&user, "name=?", name).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s mysqlStore) HasMiner(maddr address.Address) (bool, error) {
	var count int64
	err := s.db.Table("users").Where("miner=?", maddr.String()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *mysqlStore) GetMiner(maddr address.Address) (*User, error) {
	var user User
	err := s.db.Table("users").Take(&user, "miner=?", maddr.String()).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
