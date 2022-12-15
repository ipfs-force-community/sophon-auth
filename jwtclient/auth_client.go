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

//go:generate mockgen -destination=mocks/mock_auth_client.go -package=mocks github.com/filecoin-project/venus-auth/jwtclient IAuthClient

type IAuthClient interface {
	Verify(ctx context.Context, token string) (*auth.VerifyResponse, error)
	VerifyUsers(ctx context.Context, names []string) error
	HasUser(ctx context.Context, name string) (bool, error)
	GetUser(ctx context.Context, name string) (*auth.OutputUser, error)
	GetUserByMiner(ctx context.Context, miner address.Address) (*auth.OutputUser, error)
	GetUserBySigner(ctx context.Context, signer address.Address) (auth.ListUsersResponse, error)
	ListUsers(ctx context.Context, skip, limit int64, state core.UserState) (auth.ListUsersResponse, error)
	ListUsersWithMiners(ctx context.Context, skip, limit int64, state core.UserState) (auth.ListUsersResponse, error)
	GetUserRateLimit(ctx context.Context, name, id string) (auth.GetUserRateLimitResponse, error)

	MinerExistInUser(ctx context.Context, user string, miner address.Address) (bool, error)
	SignerExistInUser(ctx context.Context, user string, signer address.Address) (bool, error)

	HasMiner(ctx context.Context, miner address.Address) (bool, error)
	ListMiners(ctx context.Context, user string) (auth.ListMinerResp, error)

	HasSigner(ctx context.Context, signer address.Address) (bool, error)
	ListSigners(ctx context.Context, user string) (auth.ListSignerResp, error)
	RegisterSigners(ctx context.Context, user string, addrs []address.Address) error
	UnregisterSigners(ctx context.Context, user string, addrs []address.Address) error
}

var _ IAuthClient = (*AuthClient)(nil)

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

func (lc *AuthClient) GenerateToken(ctx context.Context, name, perm, extra string) (string, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(auth.GenTokenRequest{
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

func (lc *AuthClient) GetToken(ctx context.Context, name, token string) ([]*auth.TokenInfo, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
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

func (lc *AuthClient) Tokens(ctx context.Context, skip, limit int64) (auth.GetTokensResponse, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
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

func (lc *AuthClient) RemoveToken(ctx context.Context, token string) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(auth.RemoveTokenRequest{
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

func (lc *AuthClient) RecoverToken(ctx context.Context, token string) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(auth.RecoverTokenRequest{
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

func (lc *AuthClient) CreateUser(ctx context.Context, req *auth.CreateUserRequest) (*auth.CreateUserResponse, error) {
	resp, err := lc.cli.R().SetContext(ctx).
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
func (lc *AuthClient) UpdateUser(ctx context.Context, req *auth.UpdateUserRequest) error {
	resp, err := lc.cli.R().SetContext(ctx).
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

func (lc *AuthClient) ListUsers(ctx context.Context, skip, limit int64, state core.UserState) (auth.ListUsersResponse, error) {
	req := auth.NewListUsersRequest(skip, limit, int(state))
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
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

func (lc *AuthClient) ListUsersWithMiners(ctx context.Context, skip, limit int64, state core.UserState) (auth.ListUsersResponse, error) {
	resp, err := lc.ListUsers(ctx, skip, limit, state)
	if err != nil {
		return nil, err
	}

	for _, user := range resp {
		miners, err := lc.ListMiners(ctx, user.Name)
		if err != nil {
			log.Errorf("list user:%s miners failed:%s", user.Name, err.Error())
			continue
		}
		user.Miners = make([]*auth.OutputMiner, 0, len(miners))
		for _, val := range miners {
			user.Miners = append(user.Miners, &auth.OutputMiner{
				Miner:      val.Miner,
				User:       user.Name,
				OpenMining: val.OpenMining,
				CreatedAt:  time.Time{},
				UpdatedAt:  time.Time{},
			})
		}
	}
	return resp, nil
}

func (lc *AuthClient) VerifyUsers(ctx context.Context, names []string) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(&auth.VerifyUsersReq{Names: names}).
		SetError(&errcode.ErrMsg{}).Post("/user/verify")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUser(ctx context.Context, name string) (*auth.OutputUser, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"name": name,
	}).SetResult(&auth.OutputUser{}).SetError(&errcode.ErrMsg{}).Get("/user")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return resp.Result().(*auth.OutputUser), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) HasUser(ctx context.Context, name string) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"name": name,
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/user/has")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DeleteUser(ctx context.Context, req *auth.DeleteUserRequest) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(req).SetError(&errcode.ErrMsg{}).Post("/user/del")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) RecoverUser(ctx context.Context, req *auth.RecoverUserRequest) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(req).SetError(&errcode.ErrMsg{}).Post("/user/recover")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUserRateLimit(ctx context.Context, name, id string) (auth.GetUserRateLimitResponse, error) {
	param := make(map[string]string)
	if len(name) != 0 {
		param["name"] = name
	}
	if len(id) != 0 {
		param["id"] = id
	}

	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(param).
		SetResult(&auth.GetUserRateLimitResponse{}).
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

func (lc *AuthClient) UpsertUserRateLimit(ctx context.Context, req *auth.UpsertUserRateLimitReq) (string, error) {
	var res string
	resp, err := lc.cli.R().SetContext(ctx).SetBody(req).SetResult(&res).SetError(&errcode.ErrMsg{}).Post("/user/ratelimit/upsert")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() == http.StatusOK {
		return res, nil
	}
	return "", resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DelUserRateLimit(ctx context.Context, req *auth.DelUserRateLimitReq) (string, error) {
	var id string
	resp, err := lc.cli.R().SetContext(ctx).SetBody(req).SetResult(&id).SetError(&errcode.ErrMsg{}).Post("/user/ratelimit/del")
	if err != nil {
		return "", err
	}
	if resp.StatusCode() == http.StatusOK {
		return id, nil
	}
	return "", resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) UpsertMiner(ctx context.Context, user, miner string, openMining bool) (bool, error) {
	if _, err := address.NewFromString(miner); err != nil {
		return false, xerrors.Errorf("invalid miner address:%s", miner)
	}

	var isCreate bool
	mAddr, err := address.NewFromString(miner)
	if err != nil {
		return false, err
	}
	resp, err := lc.cli.R().SetContext(ctx).SetBody(&auth.UpsertMinerReq{Miner: mAddr, User: user, OpenMining: &openMining}).
		SetResult(&isCreate).SetError(&errcode.ErrMsg{}).Post("/user/miner/add")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return isCreate, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) HasMiner(ctx context.Context, miner address.Address) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"miner": miner.String(),
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/miner/has")
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) MinerExistInUser(ctx context.Context, user string, miner address.Address) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"miner": miner.String(),
		"user":  user,
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/user/miner/exist")
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) ListMiners(ctx context.Context, user string) (auth.ListMinerResp, error) {
	var res auth.ListMinerResp
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{"user": user}).
		SetResult(&res).SetError(&errcode.ErrMsg{}).Get("/user/miner/list")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return res, nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DelMiner(ctx context.Context, miner string) (bool, error) {
	if _, err := address.NewFromString(miner); err != nil {
		return false, xerrors.Errorf("invalid miner address:%s", miner)
	}

	var has bool
	mAddr, err := address.NewFromString(miner)
	if err != nil {
		return false, err
	}
	resp, err := lc.cli.R().SetContext(ctx).SetBody(auth.DelMinerReq{Miner: mAddr}).
		SetResult(&has).SetError(&errcode.ErrMsg{}).Post("/user/miner/del")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUserByMiner(ctx context.Context, miner address.Address) (*auth.OutputUser, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"miner": miner.String(),
	}).SetResult(&auth.OutputUser{}).SetError(&errcode.ErrMsg{}).Get("/user/miner")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return resp.Result().(*auth.OutputUser), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) RegisterSigners(ctx context.Context, user string, addrs []address.Address) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(&auth.RegisterSignersReq{Signers: addrs, User: user}).
		SetError(&errcode.ErrMsg{}).Post("/user/signer/register")
	if err != nil {
		return err
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) SignerExistInUser(ctx context.Context, user string, signer address.Address) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"signer": signer.String(),
		"user":   user,
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/user/signer/exist")
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) ListSigners(ctx context.Context, user string) (auth.ListSignerResp, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{"user": user}).
		SetResult(&auth.ListSignerResp{}).SetError(&errcode.ErrMsg{}).Get("/user/signer/list")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.ListSignerResp)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) UnregisterSigners(ctx context.Context, user string, addrs []address.Address) error {
	resp, err := lc.cli.R().SetContext(ctx).SetBody(&auth.UnregisterSignersReq{Signers: addrs, User: user}).
		SetError(&errcode.ErrMsg{}).Post("/user/signer/unregister")
	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	return resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) HasSigner(ctx context.Context, signer address.Address) (bool, error) {
	var has bool
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"signer": signer.String(),
	}).SetResult(&has).SetError(&errcode.ErrMsg{}).Get("/signer/has")
	if err != nil {
		return false, err
	}

	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) DelSigner(ctx context.Context, signer string) (bool, error) {
	var has bool
	sAddr, err := address.NewFromString(signer)
	if err != nil {
		return false, err
	}
	resp, err := lc.cli.R().SetContext(ctx).SetBody(auth.DelSignerReq{Signer: sAddr}).
		SetResult(&has).SetError(&errcode.ErrMsg{}).Post("/signer/del")
	if err != nil {
		return false, err
	}
	if resp.StatusCode() == http.StatusOK {
		return has, nil
	}
	return false, resp.Error().(*errcode.ErrMsg).Err()
}

func (lc *AuthClient) GetUserBySigner(ctx context.Context, signer address.Address) (auth.ListUsersResponse, error) {
	resp, err := lc.cli.R().SetContext(ctx).SetQueryParams(map[string]string{
		"signer": signer.String(),
	}).SetResult(&auth.ListUsersResponse{}).SetError(&errcode.ErrMsg{}).Get("/user/signer")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() == http.StatusOK {
		return *(resp.Result().(*auth.ListUsersResponse)), nil
	}
	return nil, resp.Error().(*errcode.ErrMsg).Err()
}
