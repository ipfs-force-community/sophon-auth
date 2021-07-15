package storage

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/util"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	if cnf.Debug || true {
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

	var verMajer int
	var version string

	session := db.Session(&gorm.Session{})

	// detect mysql version
	if err = session.Raw("select version();").Scan(&version).Error; err == nil {
		for _, s := range version {
			if s >= '0' && s <= '9' {
				verMajer = int(s) - 48
				break
			}
		}
	}

	// enable following code on mysql with version < 8
	if verMajer > 0 && verMajer < 8 {
		if err = session.Raw("set innodb_large_prefix = 1; set global innodb_file_format = BARRACUDA;").Error; err == nil {
			session = session.Set("gorm:table_options", "ROW_FORMAT=DYNAMIC")
		}
	}

	if err = session.AutoMigrate(&KeyPair{}, &User{}, &UserRateLimit{}); err != nil {
		return nil, err
	}

	return &mysqlStore{db: db, pkg: util.PackagePath(mysqlStore{})}, nil
}

func (s *mysqlStore) Put(kp *KeyPair) error {
	return s.db.Create(kp).Error
}

func (s mysqlStore) Delete(token Token) error {
	return s.db.Table("token").Delete(&KeyPair{}, "token=?", token).Error
}

func (s mysqlStore) Has(token Token) (bool, error) {
	var count int64
	err := s.db.Table("token").Where("token=?", token).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlStore) Get(token Token) (*KeyPair, error) {
	var kp KeyPair
	err := s.db.Table("token").Take(&kp, "token = ?", token.String()).Error

	return &kp, err
}

func (s mysqlStore) List(skip, limit int64) ([]*KeyPair, error) {
	var tokens []*KeyPair
	err := s.db.Offset(int(skip)).Limit(int(limit)).Order("name").Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *mysqlStore) UpdateToken(kp *KeyPair) error {
	columns := map[string]interface{}{
		"name":       kp.Name,
		"perm":       kp.Perm,
		"secret":     kp.Secret,
		"extra":      kp.Extra,
		"token":      kp.Token,
		"createTime": kp.CreateTime,
	}
	return s.db.Table("token").Where("token = ?", kp.Token.String()).UpdateColumns(columns).Error

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
	return s.db.Table("users").Save(user).Error
}

func (s *mysqlStore) PutUser(user *User) error {
	return s.db.Table("users").Save(user).Error
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

	return &user, err
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

func (s *mysqlStore) GetRateLimits(name string, id string) ([]*UserRateLimit, error) {
	var limits []*UserRateLimit
	tmp := s.db.Model((*UserRateLimit)(nil)).Where("name = ?", name)
	if len(id) != 0 {
		tmp = tmp.Where("id = ?", id)
	}
	return limits, tmp.Find(&limits).Error
}

func (s *mysqlStore) PutRateLimit(limit *UserRateLimit) (string, error) {
	if len(limit.Id) == 0 {
		limit.Id = uuid.NewString()
	}
	return limit.Id, s.db.Table("user_rate_limits").Save(limit).Error
}

func (s *mysqlStore) DelRateLimit(name, id string) error {
	return s.db.Table("user_rate_limits").
		Where("id = ? and name= ?", id, name).
		Delete(nil).Error
}
