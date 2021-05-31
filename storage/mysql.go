package storage

import (
	"context"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/errcode"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/filecoin-project/venus-auth/util"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

type mysqlStore struct {
	db  *sqlx.DB
	pkg string
}

func newMySQLStore(cnf *config.DBConfig) (Store, error) {
	db, err := sqlx.ConnectContext(context.Background(), "mysql", cnf.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(cnf.MaxIdleConns)
	db.SetMaxOpenConns(cnf.MaxOpenConns)
	db.SetConnMaxIdleTime(cnf.MaxIdleTime)
	db.SetConnMaxLifetime(cnf.MaxLifeTime)
	store := &mysqlStore{
		db:  db,
		pkg: util.PackagePath(mysqlStore{}),
	}
	err = store.initTable()
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *mysqlStore) Put(kp *KeyPair) error {
	res, err := s.db.NamedExec(`INSERT INTO token (token,createTime,name,perm,extra) VALUES (:token,:createTime,:name,:perm,:extra )`, kp)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "Put/NamedExec",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	count, err := res.RowsAffected()
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "Put/RowsAffected",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	if count == 0 {
		return errcode.ErrSystemExecFailed
	}
	return nil
}

func (s mysqlStore) Delete(token Token) error {
	res, err := s.db.Exec(`DELETE FROM token WHERE token = ?`, token)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "Delete/Exec",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	count, err := res.RowsAffected()
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "Delete/RowsAffected",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	if count == 0 {
		return errcode.ErrSystemExecFailed
	}
	return nil
}

func (s mysqlStore) Has(token Token) (bool, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) as count FROM token WHERE token=?`, token)
	err := row.Scan(&count)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "Has/Scan",
			},
		).Error(err)
		return false, errcode.ErrSystemExecFailed
	}
	return count > 0, nil
}

func (s mysqlStore) List(skip, limit int64) ([]*KeyPair, error) {
	arr := make([]*KeyPair, 0)
	err := s.db.Select(&arr, "SELECT * FROM token ORDER BY createTime LIMIT ?,?", skip, limit)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "List/Select",
			},
		).Error(err)
		return nil, errcode.ErrSystemExecFailed
	}
	return arr, nil
}

func (s mysqlStore) HasUser(name string) (bool, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) as count FROM users WHERE name=?`, name)
	err := row.Scan(&count)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "HasUser/Scan",
			},
		).Error(err)
		return false, errcode.ErrSystemExecFailed
	}
	return count > 0, nil
}

func (s *mysqlStore) UpdateUser(user *User) error {
	user.UpdateTime = time.Now().Local()
	res, err := s.db.NamedExec(`
		UPDATE users SET 
		miner=:miner, 
		comment=:comment, 
		state=:state, 
		stype=:stype, 
		updateTime=:updateTime
		where name=:name`,
		user)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "UpdateUser/NamedExec",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	count, err := res.RowsAffected()
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "UpdateUser/RowsAffected",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	if count == 0 {
		return errcode.ErrSystemExecFailed
	}
	return nil
}

func (s *mysqlStore) PutUser(user *User) error {
	res, err := s.db.NamedExec(`
	INSERT INTO users 
	(id, name, miner, comment, state, createTime, updateTime, stype) 
	VALUES 
	(:id,:name,:miner,:comment,:state,:createTime,:updateTime,:stype )`,
		user)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "ListUsers/NamedExec",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	count, err := res.RowsAffected()
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "ListUsers/RowsAffected",
			},
		).Error(err)
		return errcode.ErrSystemExecFailed
	}
	if count == 0 {
		return errcode.ErrSystemExecFailed
	}
	return nil
}

func (s *mysqlStore) ListUsers(skip, limit int64, state int, sourceType core.SourceType, code core.KeyCode) ([]*User, error) {
	arr := make([]*User, 0)
	query := "SELECT * FROM users "
	params := make([]interface{}, 0, 4)
	where := false
	if code&1 == 1 {
		where = true
		query += "WHERE stype=? "
		params = append(params, sourceType)
	}
	if code&2 == 2 {
		if where {
			query += "AND "
		} else {
			query += "WHERE "
		}
		query += "state=? "
		params = append(params, state)
	}
	query += "ORDER BY createTime LIMIT ?,?"
	params = append(params, skip, limit)

	err := s.db.Select(&arr, query, params)
	if err != nil {
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "ListUsers/Select",
			},
		).Error(err)
		return nil, errcode.ErrSystemExecFailed
	}
	return arr, nil
}

func (s *mysqlStore) GetUser(name string) (*User, error) {
	var user User
	err := s.db.Get(&user, "SELECT * FROM users where name=?", name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errcode.ErrDataNotExists
		}
		log.WithFields(
			log.Fields{
				"pkg":    s.pkg,
				"method": "GetUser/Get",
			},
		).Error(err)
		return nil, errcode.ErrSystemExecFailed
	}
	return &user, nil
}
