package util

import (
	"fmt"
	"os"
)

func MakeDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0o755)
			if err != nil {
				return fmt.Errorf("make dir: %w", err)
			}
		} else {
			return fmt.Errorf("stat dir: %w", err)
		}
	} else {
		if !fi.IsDir() {
			return fmt.Errorf("path %s is not a dir", path)
		}
	}
	return nil
}

func Exist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}
