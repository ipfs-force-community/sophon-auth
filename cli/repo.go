package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/mitchellh/go-homedir"
)

type Repo interface {
	GetConfig() (*config.Config, error)
	SaveConfig(*config.Config) error
	GetToken() (string, error)
	SaveToken(string) error
	GetDataDir() (string, error)
}

const (
	// DefaultConfigFile is the default config file name
	DefaultConfigFile = "config.toml"
	// DefaultDataDir is the default data directory name
	DefaultDataDir = "data"
	// DefaultTokenFile is the default token file name
	DefaultTokenFile = "token"
)

type FsRepo struct {
	repoPath string
	// configPath is the relative path to the config file from the repoPath
	configPath string
	// dataPath is the relative path to the data directory from the repoPath
	dataPath string
	// tokenPath is the relative path to the token file from the repoPath
	tokenPath string
}

func (r *FsRepo) GetConfig() (*config.Config, error) {
	path := path.Join(r.repoPath, r.configPath)
	exist, err := exist(path)
	if err != nil {
		return nil, fmt.Errorf("check config exist: %w", err)
	}
	if exist {
		cnf, err := config.DecodeConfig(path)
		if err != nil {
			return nil, fmt.Errorf("decode config: %w", err)
		}
		return cnf, nil
	}
	cnf, err := config.DefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	err = config.Cover(path, cnf)
	if err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}
	return cnf, nil
}

func (r *FsRepo) SaveConfig(cnf *config.Config) error {
	path := path.Join(r.repoPath, r.configPath)
	return config.Cover(path, cnf)
}

func (r *FsRepo) GetToken() (string, error) {
	path := path.Join(r.repoPath, r.tokenPath)
	exist, err := exist(path)
	if err != nil {
		return "", fmt.Errorf("check token exist: %w", err)
	}
	if !exist {
		return "", fmt.Errorf("token not exist")
	}
	token, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read token: %w", err)
	}

	return string(token), nil
}

func (r *FsRepo) SaveToken(token string) error {
	path := path.Join(r.repoPath, r.tokenPath)
	return os.WriteFile(path, []byte(token), os.ModePerm)
}

func (r *FsRepo) GetDataDir() (string, error) {
	ret := path.Join(r.repoPath, r.dataPath)
	err := makeDir(ret)
	if err != nil {
		return "", fmt.Errorf("make data dir: %w", err)
	}
	return ret, nil
}

func NewFsRepo(repoPath string) (Repo, error) {
	var err error
	repoPath, err = homedir.Expand(repoPath)
	if err != nil {
		return nil, fmt.Errorf("expand home dir: %w", err)
	}
	ret := &FsRepo{
		repoPath:   repoPath,
		configPath: DefaultConfigFile,
		dataPath:   DefaultDataDir,
		tokenPath:  DefaultTokenFile,
	}
	// create repo if not exist
	err = makeDir(repoPath)
	if err != nil {
		return nil, fmt.Errorf("make repo dir: %w", err)
	}
	return ret, nil
}

func makeDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
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

func exist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}
