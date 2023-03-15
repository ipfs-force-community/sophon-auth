package cli

import (
	"fmt"

	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

func GetCli(ctx *cli.Context) (*jwtclient.AuthClient, error) {
	repoPath, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return nil, err
	}

	repo, err := NewFsRepo(repoPath)
	if err != nil {
		return nil, fmt.Errorf("create repo: %w", err)
	}

	cnf, err := repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	token, err := repo.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	return jwtclient.NewAuthClient("http://localhost:"+cnf.Port, token)
}
