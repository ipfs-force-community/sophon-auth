package auth

import (
	"net/http"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/gin-gonic/gin"
)

type OAuthApp interface {
	Verify(c *gin.Context)
	GenerateToken(c *gin.Context)
	RemoveToken(c *gin.Context)
	Tokens(c *gin.Context)

	UpdateUser(c *gin.Context)
	CreateUser(c *gin.Context)
	ListUsers(c *gin.Context)
	GetMiner(c *gin.Context)
	HasMiner(c *gin.Context)
	GetUser(c *gin.Context)

	AddUserRateLimit(c *gin.Context)
	UpdateUserRateLimit(c *gin.Context)
	UpsertUserRateLimit(c *gin.Context)
	GetUserRateLimit(c *gin.Context)
	DelUserRateLimit(c *gin.Context)

	UpsertMiner(c *gin.Context)
	ListMiners(c *gin.Context)
	DelMiner(c *gin.Context)
}

type oauthApp struct {
	srv OAuthService
}

func NewOAuthApp(secret, dbPath string, cnf *config.DBConfig) (OAuthApp, error) {
	srv, err := NewOAuthService(secret, dbPath, cnf)
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
		c.Error(err) // nolint
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.AbortWithStatus(http.StatusOK)
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
	// todo check miner exit
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
	// todo check miner exit
	err := o.srv.UpdateUser(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
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

func (o *oauthApp) GetMiner(c *gin.Context) {
	req := new(GetMinerRequest)
	if err := c.ShouldBindQuery(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.GetMiner(c, req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
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
func (o *oauthApp) UpdateUserRateLimit(c *gin.Context) {

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

func (o *oauthApp) DelMiner(c *gin.Context) {
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
