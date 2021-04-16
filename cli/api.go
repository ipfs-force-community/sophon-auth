package cli

import (
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/ipfs-force-community/venus-auth/auth"
	"github.com/ipfs-force-community/venus-auth/config"
	"github.com/ipfs-force-community/venus-auth/core"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"net/http"
	"path"
	"strconv"
)

type LocalClient interface {
	GenerateToken(name, perm, extra string) (string, error)
	Tokens(pageIndex, pageSize int64) (auth.GetTokensResponse, error)
	RemoveToken(token string) error
}

type localClient struct {
	cli *resty.Client
}

type ErrMsg struct {
	Error string `json:"error"`
}

func (err *ErrMsg) Err() error {
	return errors.New(err.Error)
}

func GetCli(ctx *cli.Context) (*localClient, error) {
	p, err := homedir.Expand(ctx.String("repo"))
	if err != nil {
		return nil, xerrors.Errorf("could not expand home dir (repo): %w", err)
	}
	cnfPath, err := homedir.Expand(ctx.String("config"))
	if err != nil {
		return nil, xerrors.Errorf("could not expand home dir (config): %w", err)
	}
	if len(cnfPath) == 0 {
		cnfPath = path.Join(p, "config.toml")
	}
	cnf, err := config.DecodeConfig(cnfPath)
	if err != nil {
		return nil, xerrors.Errorf("failed to decode config err: %w", err)
	}
	return newClient(cnf.Port)
}

func newClient(port string) (*localClient, error) {
	client := resty.New().
		SetHostURL("http://localhost:"+port).
		SetHeader("Accept", "application/json")
	return &localClient{cli: client}, nil
}

func (lc *localClient) GenerateToken(name, perm, extra string) (string, error) {
	resp, err := lc.cli.R().SetBody(&auth.GenTokenRequest{
		Name:  name,
		Perm:  perm,
		Extra: extra,
	}).SetResult(&auth.GenTokenResponse{}).SetError(&ErrMsg{}).Post("/genToken")
	if err != nil {
		return core.EmptyString, err
	}
	if resp.StatusCode() == http.StatusOK {
		res := resp.Result().(*auth.GenTokenResponse)
		return res.Token, nil
	}
	return core.EmptyString, resp.Error().(*ErrMsg).Err()
}

func (lc *localClient) Tokens(skip, limit int64) (auth.GetTokensResponse, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"skip":  strconv.FormatInt(skip, 10),
		"limit": strconv.FormatInt(limit, 10),
	}).SetResult(&auth.GetTokensResponse{}).SetError(&ErrMsg{}).Get("/tokens")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.GetTokensResponse)), nil
	}
	return nil, resp.Error().(*ErrMsg).Err()
}

func (lc *localClient) RemoveToken(token string) error {
	resp, err := lc.cli.R().SetBody(&auth.RemoveTokenRequest{
		Token: token,
	}).SetError(&ErrMsg{}).Delete("/token")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*ErrMsg).Err()
}
