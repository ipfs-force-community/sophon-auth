package auth

import (
	"github.com/filecoin-project/venus-auth/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

type OAuthApp interface {
	Verify(c *gin.Context)
	GenerateToken(c *gin.Context)
	RemoveToken(c *gin.Context)
	Tokens(c *gin.Context)

	UpdateUser(c *gin.Context)
	CreateUser(c *gin.Context)
	ListUsers(c *gin.Context)
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
	c.Error(err) //nolint
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func SuccessResponse(c *gin.Context, obj interface{}) {
	c.JSON(http.StatusOK, obj)
}

func Response(c *gin.Context, err error) {
	if err != nil {
		c.Error(err) //nolint
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
			c.Error(err) //nolint
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
	if err != nil {
		BadResponse(c, err)
		return
	}
	Response(c, err)
}

func (o *oauthApp) ListUsers(c *gin.Context) {
	req := new(ListUsersRequest)
	if err := c.ShouldBind(req); err != nil {
		BadResponse(c, err)
		return
	}
	res, err := o.srv.ListUsers(c,req)
	if err != nil {
		BadResponse(c, err)
		return
	}
	SuccessResponse(c, res)
}
