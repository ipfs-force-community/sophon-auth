package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
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

	if err = session.AutoMigrate(&KeyPair{}, &User{}, &Signer{}, &UserRateLimit{}, &StoreVersion{}); err != nil {
		return nil, err
	}

	// `miners` table changes the primary key in V1.9.0. AutoMigrate will fail, so need to handle migration independently.
	if bHas := session.Migrator().HasTable(&Miner{}); !bHas {
		if err := session.Migrator().CreateTable(&Miner{}); err != nil {
			return nil, err
		}
	} else {
		if bHas := session.Migrator().HasColumn(&Miner{}, "ID"); bHas {
			if err := session.AutoMigrate(&Miner{}); err != nil {
				return nil, err
			}
		}
	}

	return &mysqlStore{db: db}, nil
}

func (s *mysqlStore) Put(kp *KeyPair) error {
	return s.db.Create(kp).Error
}

func (s mysqlStore) Delete(token Token) error {
	has, err := s.Has(token)
	if err != nil {
		return err
	}
	if !has {
		return gorm.ErrRecordNotFound
	}
	return s.db.Table("token").Where("token=?", token.String()).Update("is_deleted", core.Deleted).Error
}

func (s mysqlStore) Recover(token Token) error {
	var count int64
	err := s.db.Table("token").Where("token=? and is_deleted=?", token.String(), core.Deleted).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return s.db.Table("token").Where("token=?", token.String()).Update("is_deleted", core.NotDelete).Error
	}

	return gorm.ErrRecordNotFound
}

func (s mysqlStore) Has(token Token) (bool, error) {
	var count int64
	err := s.db.Table("token").Where("token=? and is_deleted=?", token.String(), core.NotDelete).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlStore) Get(token Token) (*KeyPair, error) {
	var kp KeyPair
	err := s.db.Table("token").Take(&kp, "token = ? and is_deleted=?", token.String(), core.NotDelete).Error

	return &kp, err
}

func (s *mysqlStore) ByName(name string) ([]*KeyPair, error) {
	var tokens []*KeyPair
	err := s.db.Find(&tokens, "name = ? and is_deleted=?", name, core.NotDelete).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s mysqlStore) List(skip, limit int64) ([]*KeyPair, error) {
	var tokens []*KeyPair
	err := s.db.Offset(int(skip)).Limit(int(limit)).Order("name").Find(&tokens, "is_deleted=?", core.NotDelete).Error
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
		"is_deleted": kp.IsDeleted,
	}
	return s.db.Table("token").Where("token = ?", kp.Token.String()).UpdateColumns(columns).Error
}

func (s mysqlStore) HasUser(name string) (bool, error) {
	var count int64
	err := s.db.Table("users").Where("name=? and is_deleted=?", name, core.NotDelete).Count(&count).Error

	return count > 0, err
}

func (s *mysqlStore) UpdateUser(user *User) error {
	return s.innerUpdateUser(s.db, user)
}

func (s *mysqlStore) innerUpdateUser(tx *gorm.DB, user *User) error {
	return tx.Table("users").Save(user).Error
}

func (s *mysqlStore) PutUser(user *User) error {
	return s.db.Table("users").Create(user).Error
}

func (s *mysqlStore) ListUsers(skip, limit int64, state core.UserState) ([]*User, error) {
	exec := s.db.Table("users")
	if state != core.UserStateUndefined {
		exec = exec.Where("state=?", state)
	}
	arr := make([]*User, 0)
	err := exec.Where("is_deleted=?", core.NotDelete).Order("createTime").Offset(int(skip)).Limit(int(limit)).Scan(&arr).Error
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (s *mysqlStore) GetUser(name string) (*User, error) {
	return s.innerGetUser(s.db, name)
}

func (s *mysqlStore) innerGetUser(tx *gorm.DB, name string) (*User, error) {
	var user User
	err := tx.Table("users").Take(&user, "name=? and is_deleted=?", name, core.NotDelete).Error

	return &user, err
}

func (s *mysqlStore) VerifyUsers(names []string) error {
	for _, name := range names {
		if _, err := s.innerGetUser(s.db, name); err != nil {
			return xerrors.Errorf("verify %s: %w", name, err)
		}
	}

	return nil
}

func (s *mysqlStore) DeleteUser(userName string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		user, err := s.innerGetUser(tx, userName)
		if err != nil {
			return err
		}

		user.IsDeleted = core.Deleted
		user.UpdateTime = time.Now()

		miners, err := s.innerListMiners(tx, userName)
		if err != nil {
			return err
		}

		minerAddrs := make([]address.Address, 0, len(miners))
		for _, miner := range miners {
			_, err := s.innerDelMiner(tx, miner.Miner.Address())
			if err != nil {
				return err
			}
			minerAddrs = append(minerAddrs, miner.Miner.Address())
		}

		// delete signer of user
		counts, err := s.innerDelSignerOfUser(tx, user.Name)
		if err != nil {
			return err
		}
		log.Infof("delete %v signers of %s", counts, userName)

		if err := s.innerUpdateUser(tx, user); err != nil {
			return err
		}

		log.Infof("delete user %s, delete miners %v", userName, minerAddrs)
		return nil
	})

	return err
}

func (s *mysqlStore) RecoverUser(userName string) error {
	var user User
	err := s.db.Table("users").Take(&user, "name=? and is_deleted=?", userName, core.Deleted).Error
	if err != nil {
		return err
	}

	user.IsDeleted = core.NotDelete
	return s.innerUpdateUser(s.db, &user)
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
	if len(name) == 0 || len(id) == 0 {
		return errors.New("user and rate-limit id is required for removing rate limit regulation")
	}

	return s.db.Table("user_rate_limits").
		Where("id = ? and name= ?", id, name).
		Delete(nil).Error
}

func (s *mysqlStore) GetUserByMiner(miner address.Address) (*User, error) {
	var user User
	if err := s.db.Model(&Miner{}).Select("users.*").
		Joins("inner join users on miners.`miner` = ? and users.`name` = miners.`user` and users.`is_deleted` = ?", storedAddress(miner), core.NotDelete).
		Scan(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *mysqlStore) UpsertMiner(maddr address.Address, userName string, openMining *bool) (bool, error) {
	var isCreate bool
	stoMiner := storedAddress(maddr)
	return isCreate, s.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Model(&user).First(&user, "name = ?", userName).Error; err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				return xerrors.Errorf("can't bind miner:%s to not exist user:%s", maddr.String(), userName)
			}
			return xerrors.Errorf("bind miner:%s to user:%s failed:%w", maddr.String(), userName, err)
		}
		var count int64
		if err := tx.Model(&Miner{}).Where("miner = ?", stoMiner).Count(&count).Error; err != nil {
			return err
		}
		// 声明了默认值的字段, 通过结构体更新数据库时gorm库会忽略零值: 0, nil, "", false 等. 可以用map或把字段定义为指针方式避免
		isCreate = count == 0
		return tx.Model(&Miner{}).
			Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "miner"}},
				UpdateAll: true,
			}).
			Create(&Miner{Miner: stoMiner, User: user.Name, OpenMining: openMining}).Error
	}, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
}

func (s mysqlStore) HasMiner(maddr address.Address) (bool, error) {
	var count int64
	if err := s.db.Table("miners").Where("miner = ? AND deleted_at IS NULL", storedAddress(maddr)).Count(&count).Error; err != nil {
		return false, nil
	}

	return count > 0, nil
}

func (s mysqlStore) MinerExistInUser(maddr address.Address, userName string) (bool, error) {
	var count int64
	if err := s.db.Table("miners").Where("miner = ? AND user = ? AND deleted_at IS NULL", storedAddress(maddr), userName).Count(&count).Error; err != nil {
		return false, nil
	}

	return count > 0, nil
}

func (s *mysqlStore) DelMiner(miner address.Address) (bool, error) {
	return s.innerDelMiner(s.db, miner)
}

func (s *mysqlStore) innerDelMiner(tx *gorm.DB, miner address.Address) (bool, error) {
	db := tx.Model((*Miner)(nil)).Delete(&Miner{}, "miner = ?", storedAddress(miner))

	return db.RowsAffected > 0, db.Error
}

func (s *mysqlStore) ListMiners(user string) ([]*Miner, error) {
	return s.innerListMiners(s.db, user)
}

func (s *mysqlStore) innerListMiners(tx *gorm.DB, user string) ([]*Miner, error) {
	var miners []*Miner
	if err := tx.Model((*Miner)(nil)).Find(&miners, "user = ?", user).Error; err != nil {
		return nil, err
	}
	return miners, nil
}

func (s *mysqlStore) RegisterSigner(addr address.Address, userName string) error {
	storedSigner := storedAddress(addr)
	return s.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Model(&user).First(&user, "`name` = ?", userName).Error; err != nil {
			if xerrors.Is(err, gorm.ErrRecordNotFound) {
				return xerrors.Errorf("can't bind signer:%s to not exist user:%s", addr.String(), userName)
			}
			return xerrors.Errorf("bind signer:%s to user:%s failed:%w", addr.String(), userName, err)
		}
		return tx.Model(&Signer{}).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "signer"}, {Name: "user"}},
				UpdateAll: true, // created_at will not be updated
			}).
			Create(&Signer{Signer: storedSigner, User: user.Name}).Error
	}, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false})
}

func (s mysqlStore) SignerExistInUser(addr address.Address, userName string) (bool, error) {
	var count int64
	if err := s.db.Table("signers").Where("`signer` = ? AND `user` = ? AND deleted_at IS NULL", storedAddress(addr), userName).Count(&count).Error; err != nil {
		return false, nil
	}

	return count > 0, nil
}

func (s *mysqlStore) innerListSigners(tx *gorm.DB, user string) ([]*Signer, error) {
	var signers []*Signer
	if err := tx.Model((*Signer)(nil)).Find(&signers, "`user` = ?", user).Error; err != nil {
		return nil, err
	}
	return signers, nil
}

func (s *mysqlStore) ListSigner(user string) ([]*Signer, error) {
	return s.innerListSigners(s.db, user)
}

func (s *mysqlStore) UnregisterSigner(addr address.Address, userName string) error {
	db := s.db.Model((*Signer)(nil)).Delete(&Signer{}, "`signer` = ? AND `user` = ?", storedAddress(addr), userName)

	return db.Error
}

func (s mysqlStore) HasSigner(addr address.Address) (bool, error) {
	var count int64
	if err := s.db.Table("signers").Where("`signer` = ? AND deleted_at IS NULL", storedAddress(addr)).Count(&count).Error; err != nil {
		return false, nil
	}

	return count > 0, nil
}

func (s *mysqlStore) DelSigner(addr address.Address) (bool, error) {
	db := s.db.Model((*Signer)(nil)).Delete(&Signer{}, "`signer` = ?", storedAddress(addr))

	return db.RowsAffected > 0, db.Error
}

func (s *mysqlStore) GetUserBySigner(addr address.Address) ([]*User, error) {
	var users []*User
	if err := s.db.Model(&Signer{}).Select("users.*").
		Joins("inner join users on signers.`signer` = ? and users.`name` = signers.`user` and users.`is_deleted` = ?", storedAddress(addr), core.NotDelete).
		Scan(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return users, nil
}

func (s *mysqlStore) innerDelSignerOfUser(tx *gorm.DB, userName string) (int64, error) {
	db := tx.Model((*Signer)(nil)).Delete(&Signer{}, "`user` = ?", userName)

	return db.RowsAffected, db.Error
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

	now := time.Now()
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

func (s *mysqlStore) MigrateToV2() error {
	// `open-mining` upgrade dependency field default value implementation
	return s.db.Model(&StoreVersion{}).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&StoreVersion{ID: 1, Version: 2}).Error
}

func (s *mysqlStore) MigrateToV3() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if bHas := tx.Migrator().HasColumn(&Miner{}, "ID"); !bHas {
			if err := tx.Exec("alter table `miners` drop primary key;").Error; err != nil {
				return err
			}

			if err := tx.Exec("alter table `miners` add column `id` bigint(20) not null auto_increment primary key first;").Error; err != nil {
				return err
			}

			// normal index to unique index
			if err := tx.Exec("alter table `miners` drop index `user_miner_idx`;").Error; err != nil {
				return err
			}

			if err := tx.Exec("alter table `miners` add unique index `miner_idx` (`miner`);").Error; err != nil {
				return err
			}
		}

		return tx.Model(&StoreVersion{}).
			Clauses(clause.OnConflict{UpdateAll: true}).
			Create(&StoreVersion{ID: 1, Version: 3}).Error
	})
}
