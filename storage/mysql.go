package storage

import (
	"context"
	"errors"
	"github.com/filecoin-project/venus-auth/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"time"
)

type mysqlStore struct {
	db *sqlx.DB
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
	return &mysqlStore{db: db}, nil
}

func (s *mysqlStore) Put(kp *KeyPair) error {
	res, err := s.db.NamedExec(`INSERT INTO token (token,createTime,name,perm,extra) VALUES (:token,:createTime,:name,:perm,:extra )`, kp)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("failed to save into db")
	}
	return nil
}

func (s mysqlStore) Delete(token Token) error {
	res, err := s.db.Exec(`DELETE FROM token WHERE token = ?`, token)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("token not match")
	}
	return nil
}

func (s mysqlStore) Has(token Token) (bool, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) as count FROM token WHERE token=?`, token)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s mysqlStore) List(skip, limit int64) ([]*KeyPair, error) {
	arr := make([]*KeyPair, 0)
	err := s.db.Select(&arr, "SELECT * FROM token ORDER BY createTime LIMIT ?,?", skip, limit)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (s mysqlStore) HasUser(user *User) (bool, error) {
	var count int64
	row := s.db.QueryRow(`SELECT COUNT(*) as count FROM users WHERE id=?`, user.Id)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *mysqlStore) UpdateUser(user *User) error {
	has, err := s.HasUser(user)
	if err != nil {
		return err
	}

	if has {
		res, err := s.db.Exec(`UPDATE users SET miner=?, comment=?, state=?, updateTime=?) where id=?`, user.Miner, user.Comment, user.State, time.Now().Unix(), user.Id)
		if err != nil {
			return err
		}
		count, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.New("failed to update user")
		}
		return nil
	} else {
		return s.addUser(user)
	}
}

func (s *mysqlStore) addUser(user *User) error {
	createTime := time.Now().Unix()
	res, err := s.db.Exec(`INSERT INTO users (id,name,miner,comment,state, createTime, updateTime) VALUES (?,?,?,?,?,? )`,
		user.Id, user.Name, user.Miner, user.Comment, user.State, createTime, createTime)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("failed to save into db")
	}
	return nil
}

func (s *mysqlStore) ListUser(skip, limit int64) ([]*User, error) {
	arr := make([]*User, 0)
	err := s.db.Select(&arr, "SELECT * FROM users ORDER BY createTime LIMIT ?,?", skip, limit)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

func (s *mysqlStore) GetUser(name string) (*User, error) {
	var user User
	err := s.db.Select(&user, "SELECT * FROM users where name=?", name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
