package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/xerrors"

	"github.com/gin-gonic/gin"

	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/core"
)

// DefaultAdminToken is the default admin token which is for local client user
const DefaultAdminTokenName = "defaultLocalToken"

type OAuthApp interface {
	verify(token string) (*JWTPayload, error)
	GetDefaultAdminToken() (string, error)

	Verify(c *gin.Context)
	GenerateToken(c *gin.Context)
	RemoveToken(c *gin.Context)
	RecoverToken(c *gin.Context)
	Tokens(c *gin.Context)
	GetToken(c *gin.Context)

	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	VerifyUsers(c *gin.Context)
	ListUsers(c *gin.Context)
	HasUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	RecoverUser(c *gin.Context)

	AddUserRateLimit(c *gin.Context)
	UpsertUserRateLimit(c *gin.Context)
	GetUserRateLimit(c *gin.Context)
	DelUserRateLimit(c *gin.Context)

	UpsertMiner(c *gin.Context)
	HasMiner(c *gin.Context)
	MinerExistInUser(c *gin.Context)
	ListMiners(c *gin.Context)
	DeleteMiner(c *gin.Context)
	GetUserByMiner(c *gin.Context)

	RegisterSigners(c *gin.Context)
	SignerExistInUser(c *gin.Context)
	ListSigner(c *gin.Context)
	UnregisterSigners(c *gin.Context)
	HasSigner(c *gin.Context)
	DelSigner(c *gin.Context)
	GetUserBySigner(c *gin.Context)
}

type oauthApp struct {
	srv OAuthService
}

func NewOAuthApp(dbPath string, cnf *config.DBConfig) (OAuthApp, error) {
	srv, err := NewOAuthService(dbPath, cnf)
	if err != nil {
		return nil, err
	}
	return &oauthApp{
		srv: srv,
	}, nil
}

func BadResponse(c *gin.Context, err error) {
	c.Error(err) // nolint
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func SuccessResponse(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

func Response(c *gin.Context, err error) {
	if err != nil {
		BadResponse(c, err)
		return
	}
	c.AbortWithStatus(http.StatusOK)
}

// verify only called by inner, so use readCtx constant to bypass perm check
func (o *oauthApp) verify(token string) (*JWTPayload, error) {
	return o.srv.Verify(core.CtxWithPerm(context.Background(), core.PermRead), token)
}

func (o *oauthApp) GetDefaultAdminToken() (string, error) {
	adminCtx := core.CtxWithPerm(context.Background(), core.PermAdmin)
	// if not found, create one
	token, err := o.srv.GetTokenByName(adminCtx, DefaultAdminTokenName)
	if err != nil {
		return "", err
	}
	for _, t := range token {
		if t.Perm == core.PermAdmin {
			return t.Token, nil
		}
	}
	// create one
	_, err = o.srv.CreateUser(adminCtx, &CreateUserRequest{Name: DefaultAdminTokenName})
	if err != nil {
		return "", fmt.Errorf("create default user for admin token: %w", err)
	}
	ret, err := o.srv.GenerateToken(adminCtx, &JWTPayload{
		Name: DefaultAdminTokenName,
		Perm: core.PermAdmin,
	})
	if err != nil {
		return "", fmt.Errorf("create default admin token: %w", err)
	}

	return ret, nil
}

func (o *oauthApp) Verify(c *gin.Context) {
	req := new(VerifyRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.Verify(c, req.Token)
	if err != nil {
		if err == ErrorNonRegisteredToken || err == ErrorVerificationFailed {
			c.Error(err) // nolint
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		BadResponse(c, err)
		return
	}
	c.Set("name", res.Name)
	SuccessResponse(c, res)
}

func (o *oauthApp) GenerateToken(c *gin.Context) {
	req := new(GenTokenRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.GenerateToken(c, &JWTPayload{
		Name:  req.Name,
		Perm:  req.Perm,
		Extra: req.Extra,
	})
	if err != nil {
		BadResponse(c, err)
		return
	}
	output := &GenTokenResponse{
		Token: res,
	}
	SuccessResponse(c, output)
}

func (o *oauthApp) RemoveToken(c *gin.Context) {
	req := new(RemoveTokenRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.RemoveToken(c, req.Token)
	Response(c, err)
}

func (o *oauthApp) RecoverToken(c *gin.Context) {
	req := new(RecoverTokenRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.RecoverToken(c, req.Token)
	Response(c, err)
}

func (o *oauthApp) GetToken(c *gin.Context) {
	req := new(GetTokenRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	if len(req.Name) == 0 && len(req.Token) == 0 {
		BadResponse(c, xerrors.Errorf("`name` and `token` both empty"))
		return
	}
	var res []*TokenInfo
	if len(req.Token) > 0 {
		info, err := o.srv.GetToken(c, req.Token)
		if err != nil {
			BadResponse(c, err)
			return
		}
		res = append(res, info)
	} else {
		var err error
		res, err = o.srv.GetTokenByName(c, req.Name)
		if err != nil {
			BadResponse(c, err)
			return
		}
	}

	SuccessResponse(c, res)
}

func (o *oauthApp) Tokens(c *gin.Context) {
	req := new(GetTokensRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.Tokens(c, req.GetSkip(), req.GetLimit())
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) CreateUser(c *gin.Context) {
	req := new(CreateUserRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}

	res, err := o.srv.CreateUser(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) UpdateUser(c *gin.Context) {
	req := new(UpdateUserRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.UpdateUser(c, req)
	Response(c, err)
}

func (o *oauthApp) ListUsers(c *gin.Context) {
	req := new(ListUsersRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}

	res, err := o.srv.ListUsers(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) GetUserByMiner(c *gin.Context) {
	req := new(GetUserByMinerRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.GetUserByMiner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) GetUser(c *gin.Context) {
	req := new(GetUserRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.GetUser(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) VerifyUsers(c *gin.Context) {
	req := new(VerifyUsersReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}

	err := o.srv.VerifyUsers(c, req)
	Response(c, err)
}

func (o *oauthApp) HasUser(c *gin.Context) {
	req := new(HasUserRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	has, err := o.srv.HasUser(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, has)
}

func (o *oauthApp) DeleteUser(c *gin.Context) {
	req := new(DeleteUserRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.DeleteUser(c, req)
	Response(c, err)
}

func (o *oauthApp) RecoverUser(c *gin.Context) {
	req := new(RecoverUserRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.RecoverUser(c, req)
	Response(c, err)
}

func (o *oauthApp) AddUserRateLimit(c *gin.Context) {
	req := new(UpsertUserRateLimitReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}

	res, err := o.srv.UpsertUserRateLimit(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) UpsertUserRateLimit(c *gin.Context) {
	req := new(UpsertUserRateLimitReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}

	res, err := o.srv.UpsertUserRateLimit(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) GetUserRateLimit(c *gin.Context) {
	req := new(GetUserRateLimitsReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}

	res, err := o.srv.GetUserRateLimits(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) DelUserRateLimit(c *gin.Context) {
	req := new(DelUserRateLimitReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.DelUserRateLimit(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, req.Id)
}

func (o *oauthApp) UpsertMiner(c *gin.Context) {
	req := new(UpsertMinerReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	isCreate, err := o.srv.UpsertMiner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, isCreate)
}

func (o *oauthApp) HasMiner(c *gin.Context) {
	req := new(HasMinerRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.HasMiner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) MinerExistInUser(c *gin.Context) {
	req := new(MinerExistInUserRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.MinerExistInUser(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) ListMiners(c *gin.Context) {
	req := new(ListMinerReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.ListMiners(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) DeleteMiner(c *gin.Context) {
	req := new(DelMinerReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
	}
	res, err := o.srv.DelMiner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) RegisterSigners(c *gin.Context) {
	req := new(RegisterSignersReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.RegisterSigners(c, req)
	Response(c, err)
}

func (o *oauthApp) SignerExistInUser(c *gin.Context) {
	req := new(SignerExistInUserReq)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.SignerExistInUser(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) ListSigner(c *gin.Context) {
	req := new(ListSignerReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.ListSigner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) UnregisterSigners(c *gin.Context) {
	req := new(UnregisterSignersReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	err := o.srv.UnregisterSigners(c, req)
	Response(c, err)
}

func (o *oauthApp) HasSigner(c *gin.Context) {
	req := new(HasSignerReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
	}
	res, err := o.srv.HasSigner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) DelSigner(c *gin.Context) {
	req := new(DelSignerReq)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
	}
	res, err := o.srv.DelSigner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}

func (o *oauthApp) GetUserBySigner(c *gin.Context) {
	req := new(GetUserBySignerReq)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.GetUserBySigner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}
