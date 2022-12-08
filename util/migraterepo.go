package util

import (
	"fmt"
	"os"
	"time"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/mitchellh/go-homedir"
)

func MigrateRepo(currRepo, oldRepo, newRepo string) (string, error) {
	var err error
	newRepo, err = homedir.Expand(newRepo)
	if err != nil {
		return "", err
	}
	exist, err := config.Exist(newRepo)
	if err != nil {
		return "", err
	}
	if exist {
		return newRepo, nil
	}

	oldRepo, err = homedir.Expand(oldRepo)
	if err != nil {
		return "", err
	}
	exist, err = config.Exist(oldRepo)
	if err != nil {
		return "", err
	}
	if !exist {
		return currRepo, nil
	}

	now := time.Now()
	fmt.Printf("start move %v to %v\n", oldRepo, newRepo)
	if err := os.Rename(oldRepo, newRepo); err != nil {
		return "", err
	}
	fmt.Printf("end move repo took %v\n", time.Since(now))

	return newRepo, nil
}
