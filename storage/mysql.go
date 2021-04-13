package storage

import (
	"context"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ipfs-force-community/venus-auth/config"
	"github.com/jmoiron/sqlx"
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
	res, err := s.db.NamedExec(`INSERT INTO token (token,createTime) VALUES (:token,:createTime )`, kp)
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
