package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/util"
)

type Repo interface {
	GetConfig() (*config.Config, error)
	SaveConfig(*config.Config) error
	GetToken() (string, error)
	SaveToken(string) error
	GetDataDir() string
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
	cnf, err := config.DecodeConfig(path)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return cnf, nil
}

func (r *FsRepo) SaveConfig(cnf *config.Config) error {
	path := path.Join(r.repoPath, r.configPath)
	return config.Cover(path, cnf)
}

func (r *FsRepo) GetToken() (string, error) {
	path := path.Join(r.repoPath, r.tokenPath)
	exist, err := util.Exist(path)
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

func (r *FsRepo) GetDataDir() string {
	return path.Join(r.repoPath, r.dataPath)
}

func (r *FsRepo) init() error {
	exist, err := util.Exist(r.repoPath)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	// create repo if not exist
	err = util.MakeDir(r.repoPath)
	if err != nil {
		return fmt.Errorf("make repo dir: %w", err)
	}

	err = util.MakeDir(r.GetDataDir())
	if err != nil {
		return fmt.Errorf("make data dir: %w", err)
	}

	return nil
}

func NewFsRepo(repoPath string) (Repo, error) {
	ret := &FsRepo{
		repoPath:   repoPath,
		configPath: DefaultConfigFile,
		dataPath:   DefaultDataDir,
		tokenPath:  DefaultTokenFile,
	}

	return ret, ret.init()
}
