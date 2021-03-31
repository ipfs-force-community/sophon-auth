package db

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var db Database

func TestMain(m *testing.M) {
	path, err := ioutil.TempDir("", "badger")
	defer os.RemoveAll(path)
	if err != nil {
		os.Exit(2)
		return
	}
	db, err = Open(path)
	if err != nil {
		os.Exit(2)
		return
	}
	m.Run()
}

func TestDB(t *testing.T) {
	var (
		key = []byte("auth")
		val = []byte("FileCoin")
	)
	err := db.Put(key, val)
	if err != nil {
		t.Fatalf("put failed :%s", err)
	}
	valGet, err := db.Get(key)
	if err != nil {
		t.Fatalf("get failed :%s", err)
	}
	assert.Equal(t, val, valGet)
}
