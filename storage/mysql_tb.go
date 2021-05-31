package storage

import (
	"fmt"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strings"
)

var tables = map[string]string{
	"token": tokenSchema,
	"users": userSchema,
}

func (s *mysqlStore) initTable() error {
	for tb, schema := range tables {
		exists, err := checkTableExist(s.db, tb)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		_, err = s.db.Exec(schema)
		if err != nil {
			return err
		}
	}
	return nil
}

type Table struct {
	Table string `json:"table" db:"Table"`
	DDL   string `json:"ddl" db:"Create Table"`
}

func checkTableExist(db *sqlx.DB, table string) (bool, error) {
	tb := new(Table)
	err := db.Get(tb, "SHOW CREATE TABLE "+table)
	if err != nil {
		mysqlErr := err.(*mysql.MySQLError)
		if mysqlErr.Number == 1146 {
			return false, nil
		}
		return false, err
	}
	// check the table is up to date
	if !compareDDL(tb.DDL, tables[table]) {
		log.WithFields(log.Fields{
			"current table": tb.Table,
		}).Warn(strings.ReplaceAll(tb.DDL, "`", ""))
		log.WithFields(log.Fields{
			"newest table": tb.Table,
		}).Warn(tables[table])
		return false, fmt.Errorf("table %s out of date", table)
	}
	return true, nil
}

func compareDDL(left, right string) bool {
	leftCp := compressionDDL(left)
	rightCp := compressionDDL(right)
	return leftCp == rightCp
}

func compressionDDL(ddl string) string {
	ddl = strings.ReplaceAll(ddl, "`", "")
	ddl = strings.ReplaceAll(ddl, "\n", "")
	ddl = strings.ReplaceAll(ddl, "\t", "")
	ddl = strings.ReplaceAll(ddl, " ", "")
	ddl = strings.ReplaceAll(ddl, ";", "")
	return ddl
}

const tokenSchema = `
CREATE TABLE token (
  name varchar(50) NOT NULL,
  token varchar(512) NOT NULL,
  createTime datetime NOT NULL,
  perm varchar(50) NOT NULL,
  extra varchar(255) DEFAULT NULL,
  UNIQUE KEY token_token_IDX (token) USING HASH
) ENGINE=InnoDB
  DEFAULT CHARSET = utf8;
`

const userSchema = `
CREATE TABLE users (
  id varchar(255) NOT NULL,
  name varchar(255) NOT NULL,
  miner varchar(255) NOT NULL,
  state tinyint(4) NOT NULL DEFAULT '0',
  comment varchar(255) NOT NULL,
  createTime datetime NOT NULL,
  updateTime datetime NOT NULL,
  stype tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (id),
  UNIQUE KEY users_name_IDX (name) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

`
