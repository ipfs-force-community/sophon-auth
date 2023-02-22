package cli

import (
	"fmt"

	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/urfave/cli/v2"
)

// nolint
func GetCli(ctx *cli.Context) (*jwtclient.AuthClient, error) {
	repo, err := NewFsRepo(ctx.String("repo"))
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
