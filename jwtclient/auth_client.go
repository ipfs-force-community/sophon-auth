package jwtclient

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/errcode"
)

type AuthClient struct {
	cli *resty.Client
}

func NewAuthClient(url string) (*AuthClient, error) {
	client := resty.New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")
	return &AuthClient{cli: client}, nil
}

func (lc *AuthClient) Verify(ctx context.Context, token string) (*auth.VerifyResponse, error) {
	ctx, span := trace.StartSpan(ctx, "AuthClient.verify",
		func(so *trace.StartOptions) { so.Sampler = trace.AlwaysSample() })
	defer span.End()

	resp, err := lc.cli.R().SetContext(ctx).
		SetBody(auth.VerifyRequest{Token: token}).
		SetResult(&auth.VerifyResponse{}).Post("/verify")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusOK {
		res := resp.Result().(*auth.VerifyResponse)
		span.AddAttributes(trace.StringAttribute("Account", res.Name))
		return res, err
	}

	span.SetStatus(trace.Status{
		Code:    trace.StatusCodeUnauthenticated,
		Message: string(resp.Body()),
	})

	return nil, fmt.Errorf("response code is : %d, msg:%s", resp.StatusCode(), resp.Body())
}

func (lc *AuthClient) GenerateToken(name, perm, extra string) (string, error) {
	resp, err := lc.cli.R().SetBody(auth.GenTokenRequest{
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

func (lc *AuthClient) GetToken(name, token string) ([]*auth.TokenInfo, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"name":  name,
		"token": token,
	}).SetResult(&[]*auth.TokenInfo{}).SetError(&errcode.ErrMsg{}).Get("/token")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*[]*auth.TokenInfo)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) Tokens(skip, limit int64) (auth.GetTokensResponse, error) {
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

func (lc *AuthClient) RemoveToken(token string) error {
	resp, err := lc.cli.R().SetBody(auth.RemoveTokenRequest{
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

func (lc *AuthClient) RecoverToken(token string) error {
	resp, err := lc.cli.R().SetBody(auth.RecoverTokenRequest{
		Token: token,
	}).SetError(&errcode.ErrMsg{}).Post("/recoverToken")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) CreateUser(req *auth.CreateUserRequest) (*auth.CreateUserResponse, error) {
	resp, err := lc.cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
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

// UpdateUser
func (lc *AuthClient) UpdateUser(req *auth.UpdateUserRequest) error {
	resp, err := lc.cli.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).SetError(&errcode.ErrMsg{}).Post("/user/update")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) ListUsers(req *auth.ListUsersRequest) (auth.ListUsersResponse, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"skip":  strconv.FormatInt(req.Skip, 10),
		"limit": strconv.FormatInt(req.Limit, 10),
		"state": strconv.Itoa(req.State),
	}).SetResult(&auth.ListUsersResponse{}).SetError(&errcode.ErrMsg{}).Get("/user/list")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.ListUsersResponse)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) ListUsersWithMiners(req *auth.ListUsersRequest) (auth.ListUsersResponse, error) {
	resp, err := lc.ListUsers(req)
	if err != nil {
		return nil, err
	}

	for _, user := range resp {
		miners, err := lc.ListMiners(user.Name)
		if err != nil {
			log.Errorf("list user:%s miners failed:%s", user.Name, err.Error())
			continue
		}
		user.Miners = make([]*auth.OutputMiner, 0, len(miners))
		for _, val := range miners {
			addr, err := address.NewFromString(val.Miner)
			if err != nil {
				log.Errorf("invalid user:%s miner:%s, %s", user.Name, val.Miner, err.Error())
				continue
			}
			user.Miners = append(user.Miners, &auth.OutputMiner{
				Miner:      addr.String(),
				User:       user.Name,
				OpenMining: val.OpenMining,
				CreatedAt:  time.Time{},
				UpdatedAt:  time.Time{},
			})
		}
	}
	return resp, nil
}

func (lc *AuthClient) GetUser(req *auth.GetUserRequest) (*auth.OutputUser, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"name": req.Name,
	}).SetResult(&auth.OutputUser{}).SetError(&errcode.ErrMsg{}).Get("/user")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return resp.Result().(*auth.OutputUser), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUserByMiner(req *auth.GetUserByMinerRequest) (*auth.OutputUser, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"miner": req.Miner,
	}).SetResult(&auth.OutputUser{}).SetError(&errcode.ErrMsg{}).Get("/user/miner")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return resp.Result().(*auth.OutputUser), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUserBySigner(req *auth.GetUserBySignerRequest) (*auth.OutputUser, error) {
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"signer": req.Signer,
	}).SetResult(&auth.OutputUser{}).SetError(&errcode.ErrMsg{}).Get("/user/signer")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return resp.Result().(*auth.OutputUser), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) HasUser(req *auth.HasUserRequest) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"name": req.Name,
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/user/has")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*bool)), nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DeleteUser(req *auth.DeleteUserRequest) error {
	resp, err := lc.cli.R().SetBody(req).SetError(&errcode.ErrMsg{}).Post("/user/del")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) RecoverUser(req *auth.RecoverUserRequest) error {
	resp, err := lc.cli.R().SetBody(req).SetError(&errcode.ErrMsg{}).Post("/user/recover")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUserRateLimit(name, id string) (auth.GetUserRateLimitResponse, error) {
	var res auth.GetUserRateLimitResponse
	param := make(map[string]string)
	if len(name) != 0 {
		param["name"] = name
	}
	if len(id) != 0 {
		param["id"] = id
	}
	resp, err := lc.cli.R().SetQueryParams(param).
		SetResult(&res).
		SetError(&errcode.ErrMsg{}).
		Get("/user/ratelimit")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.GetUserRateLimitResponse)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) UpsertUserRateLimit(req *auth.UpsertUserRateLimitReq) (string, error) {
	var res string
	resp, err := lc.cli.R().SetBody(req).SetResult(&res).SetError(&errcode.ErrMsg{}).Post("/user/ratelimit/upsert")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*string)), nil
	}
	return "", resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DelUserRateLimit(req *auth.DelUserRateLimitReq) (string, error) {
	var id string
	resp, err := lc.cli.R().SetBody(req).SetResult(&id).SetError(&errcode.ErrMsg{}).Post("/user/ratelimit/del")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*string)), nil
	}
	return "", resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) UpsertMiner(user, miner string, openMining bool) (bool, error) {
	if _, err := address.NewFromString(miner); err != nil {
		return false, xerrors.Errorf("invalid miner address:%s", miner)
	}
	var isCreate bool
	resp, err := lc.cli.R().SetBody(&auth.UpsertMinerReq{Miner: miner, User: user, OpenMining: openMining}).
		SetResult(&isCreate).SetError(&errcode.ErrMsg{}).Post("/user/miner/add")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*bool)), nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) HasMiner(req *auth.HasMinerRequest) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"miner": req.Miner,
		"user":  req.User,
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/user/miner/has")
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return *resp.Result().(*bool), nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) ListMiners(user string) (auth.ListMinerResp, error) {
	var res auth.ListMinerResp
	resp, err := lc.cli.R().SetQueryParams(map[string]string{"user": user}).
		SetResult(&res).SetError(&errcode.ErrMsg{}).Get("/user/miner/list")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.ListMinerResp)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DelMiner(miner string) (bool, error) {
	if _, err := address.NewFromString(miner); err != nil {
		return false, xerrors.Errorf("invalid miner address:%s", miner)
	}

	var res bool
	resp, err := lc.cli.R().SetBody(auth.DelMinerReq{Miner: miner}).
		SetResult(&res).SetError(&errcode.ErrMsg{}).Post("/user/miner/del")
	if err != nil {
		return res, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*bool)), nil
	}
	return res, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) UpsertSigner(user, addr string) (bool, error) {
	if _, err := address.NewFromString(addr); err != nil {
		return false, xerrors.Errorf("invalid signer address:%s", addr)
	}
	var isCreate bool
	resp, err := lc.cli.R().SetBody(&auth.UpsertSignerReq{Signer: addr, User: user}).
		SetResult(&isCreate).SetError(&errcode.ErrMsg{}).Post("/user/signer/add")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*bool)), nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) HasSigner(req *auth.HasSignerRequest) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetQueryParams(map[string]string{
		"signer": req.Signer,
		"user":   req.User,
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/user/signer/has")
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return *resp.Result().(*bool), nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) ListSigners(user string) (auth.ListSignerResp, error) {
	var res auth.ListSignerResp
	resp, err := lc.cli.R().SetQueryParams(map[string]string{"user": user}).
		SetResult(&res).SetError(&errcode.ErrMsg{}).Get("/user/signer/list")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.ListSignerResp)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DelSigner(signer string) (bool, error) {
	if _, err := address.NewFromString(signer); err != nil {
		return false, xerrors.Errorf("invalid signer address:%s", signer)
	}

	var res bool
	resp, err := lc.cli.R().SetBody(auth.DelSignerReq{Signer: signer}).
		SetResult(&res).SetError(&errcode.ErrMsg{}).Post("/user/signer/del")
	if err != nil {
		return res, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*bool)), nil
	}
	return res, resp.Error().(*errcode.ErrMsg).Err()
}
