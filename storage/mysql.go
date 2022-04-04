package storage

import (
	"database/sql"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mysqlStore struct {
	db *gorm.DB
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

	if err = session.AutoMigrate(&KeyPair{}, &User{}, &Miner{}, &UserRateLimit{}, &StoreVersion{}); err != nil {
		return nil, err
	}

	return &mysqlStore{db: db}, nil
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
	if err := s.db.Table("miners").Where("miner = ?", storedAddress(maddr)).Count(&count).Error; err != nil {
		return false, nil
	}
	return count > 0, nil
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

func (s *mysqlStore) GetUserByMiner(miner address.Address) (*User, error) {
	var users []*User
	if err := s.db.Model(&Miner{}).Select("users.*").
		Joins("inner join users on miners.miner = ? and users.name = miners.user", storedAddress(miner)).
		Scan(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return users[0], nil
}

func (s *mysqlStore) UpsertMiner(miner address.Address, userName string) (bool, error) {
	var isCreate bool
	stoMiner := storedAddress(miner)
	return isCreate, s.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Model(&user).First(&user, "name = ?", userName).Error; err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				return xerrors.Errorf("can't bind miner:%s to not exist user:%s", miner.String(), userName)
			}
			return xerrors.Errorf("bind miner:%s to user:%s failed:%w", miner.String(), userName, err)
		}
		var count int64
		if err := tx.Model(&Miner{}).Where("miner = ?", stoMiner).Count(&count).Error; err != nil {
			return err
		}
		isCreate = count > 0
		return tx.Model(&Miner{}).
			Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "miner"}}, UpdateAll: true}).
			Create(&Miner{Miner: stoMiner, User: user.Name}).Error
	}, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
}

func (s *mysqlStore) DelMiner(miner address.Address) (bool, error) {
	db := s.db.Model((*Miner)(nil)).Delete(&Miner{}, "miner = ?", storedAddress(miner))
	return db.RowsAffected > 0, db.Error
}

func (s *mysqlStore) ListMiners(user string) ([]*Miner, error) {
	var miners []*Miner
	if err := s.db.Model((*Miner)(nil)).Find(&miners, "user = ?", user).Error; err != nil {
		return nil, err
	}
	return miners, nil
}

func (s *mysqlStore) Version() (uint64, error) {
	var v StoreVersion
	if err := s.db.Model(&StoreVersion{}).First(&v).Error; err != nil {
		if xerrors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return v.Version, nil
}

func (s *mysqlStore) MigrateToV1() error {
	arr := make([]*struct {
		User  string `gorm:"column:name"`
		Miner string
	}, 0)
	if err := s.db.Table("users").Scan(&arr).Error; err != nil {
		return err
	}

	var now = time.Now()
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, u := range arr {
			maddr, err := address.NewFromString(u.Miner)
			if err != nil || maddr.Empty() {
				log.Warnf("won't migrate miner:%s, invalid miner address", u.Miner)
				continue
			}
			if err := tx.Model(&Miner{}).
				Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "miner"}}, UpdateAll: true}).
				Create(&Miner{
					Miner: storedAddress(maddr),
					User:  u.User,
					OrmTimestamp: OrmTimestamp{
						CreatedAt: now,
						UpdatedAt: now,
					},
				}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&StoreVersion{}).
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(&StoreVersion{ID: 1, Version: 1}).Error
	})
}
