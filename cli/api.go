package cli

import (
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/errcode"
	"github.com/go-resty/resty/v2"
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
	}).SetResult(&auth.GenTokenResponse{}).SetError(&errcode.ErrMsg{}).Post("/genToken")
	if err != nil {
		return core.EmptyString, err
	}
	if resp.StatusCode() == http.StatusOK {
		res := resp.Result().(*auth.GenTokenResponse)
		return res.Token, nil
	}
	return core.EmptyString, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *localClient) Tokens(skip, limit int64) (auth.GetTokensResponse, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"skip":  strconv.FormatInt(skip, 10),
		"limit": strconv.FormatInt(limit, 10),
	}).SetResult(&auth.GetTokensResponse{}).SetError(&errcode.ErrMsg{}).Get("/tokens")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.GetTokensResponse)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *localClient) RemoveToken(token string) error {
	resp, err := lc.cli.R().SetBody(&auth.RemoveTokenRequest{
		Token: token,
	}).SetError(&errcode.ErrMsg{}).Delete("/token")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *localClient) CreateUser(req *auth.CreateUserRequest) (*auth.CreateUserResponse, error) {
	resp, err := lc.cli.R().SetBody(req).
		SetResult(&auth.CreateUserResponse{}).
		SetError(&errcode.ErrMsg{}).
		Put("/user/new")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return resp.Result().(*auth.CreateUserResponse), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}
func (lc *localClient) UpdateUser(req *auth.UpdateUserRequest) error {
	resp, err := lc.cli.R().SetBody(req).SetError(&errcode.ErrMsg{}).Post("/user/update")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *localClient) ListUsers(req *auth.ListUsersRequest) (auth.ListUsersResponse, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"skip":       strconv.FormatInt(req.Skip, 10),
		"limit":      strconv.FormatInt(req.Limit, 10),
		"sourceType": strconv.Itoa(req.SourceType),
		"state":      strconv.Itoa(req.State),
		"keySum":     strconv.Itoa(req.KeySum),
	}).SetResult(&auth.ListUsersResponse{}).SetError(&errcode.ErrMsg{}).Get("/user/list")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.ListUsersResponse)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}
