package storage

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gotest.tools/assert"
	"os"
	"testing"
)

var store *mysqlStore

func TestMain(m *testing.M) {
	if os.Getenv("CI") == "test" {
		os.Exit(0)
	}
	dsn := "root:111111@(localhost:3306)/venus-auth?parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_general_ci"
	db, err := sqlx.ConnectContext(context.Background(), "mysql", dsn)
	if err != nil {
		os.Exit(1)
	}
	store = &mysqlStore{db: db}
	m.Run()
}

func TestInitTable(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	err := store.initTable()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCompressionDDL(t *testing.T) {
	str := "CREATE TABLE `users` (\n  `id` varchar(255) NOT NULL,\n  `name` varchar(255) DEFAULT NULL,\n  `miner` varchar(255) DEFAULT NULL,\n  `comment` varchar(255) DEFAULT NULL,\n  `createtime` timestamp NULL DEFAULT NULL,\n  `upadtetime` timestamp NULL DEFAULT NULL,\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8  table=users"
	str = compressionDDL(str)
	assert.Equal(t, str, "CREATETABLEusers(idvarchar(255)NOTNULL,namevarchar(255)DEFAULTNULL,minervarchar(255)DEFAULTNULL,commentvarchar(255)DEFAULTNULL,createtimetimestampNULLDEFAULTNULL,upadtetimetimestampNULLDEFAULTNULL,PRIMARYKEY(id))ENGINE=InnoDBDEFAULTCHARSET=utf8table=users")
}

func TestCheckTableExist(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	tbName := "test"
	_, err := store.db.Exec(testDelSchema)
	if err != nil {
		t.Fatal(err)
	}
	exists, err := checkTableExist(store.db, tbName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, exists, false)
	res, err := store.db.Exec(testSchema)
	if err != nil {
		t.Fatal(err)
	}
	row, err := res.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, row, int64(0))
	exists, err = checkTableExist(store.db, tbName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, exists, true)
}

const testSchema = `
CREATE TABLE test (
     name varchar(50) NOT NULL
) ENGINE=InnoDB
  DEFAULT CHARSET = utf8
  COLLATE = utf8_general_ci;
`

const testDelSchema = `
DROP TABLE IF EXISTS test;
`
