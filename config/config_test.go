package config

import (
	"testing"
)

func TestDecodeConfig(t *testing.T) {
	path := "./config.toml"
	cnf, err := DecodeConfig(path)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	t.Log(cnf)
}

func TestSafeWriteConfig(t *testing.T) {
	path := "./config.toml"
	cnf, err := DecodeConfig(path)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	p2 := "./safeConfig.toml"
	// t.Log(cnf.API, cnf.DB)
	err = Cover(p2, cnf)
	if err != nil {
		t.Fatal(err)
	}
}
