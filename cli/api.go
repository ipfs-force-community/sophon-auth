package cli

import (
	"fmt"

	"github.com/ipfs-force-community/sophon-auth/jwtclient"
	"github.com/ipfs-force-community/sophon-auth/util"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

func GetCli(ctx *cli.Context) (*jwtclient.AuthClient, error) {
	repoPath, err := GetRepoPath(ctx)
	if err != nil {
		return nil, err
	}

	repo, err := NewFsRepo(repoPath)
	if err != nil {
		return nil, fmt.Errorf("create repo: %w", err)
	}

	token, err := repo.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	var listen string
	if ctx.IsSet("listen") {
		listen = ctx.String("listen")
	} else {
		cnf, err := repo.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("get config: %w", err)
		}
		listen = cnf.Listen
	}

	return jwtclient.NewAuthClient("http://"+listen, token)
}

func GetRepoPath(ctx *cli.Context) (string, error) {
	repoPath, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return "", err
	}

	exist, err := util.Exist(repoPath)
	if err != nil {
		return "", fmt.Errorf("check repo exist: %w", err)
	}

	// todo: rm compatibility for repo when appropriate
	if !exist {
		deprecatedRepoPath, err := homedir.Expand("~/.venus-auth")
		if err != nil {
			return "", fmt.Errorf("expand deprecated home dir: %w", err)
		}

		deprecatedRepoPathExist, err := util.Exist(deprecatedRepoPath)
		if err != nil {
			return "", fmt.Errorf("check deprecated repo exist: %w", err)
		}
		if deprecatedRepoPathExist {
			fmt.Printf("[WARM]: repo path %s is deprecated, please transfer to %s instead\n", deprecatedRepoPath, repoPath)
			repoPath = deprecatedRepoPath
		}
	}

	return repoPath, nil
}
