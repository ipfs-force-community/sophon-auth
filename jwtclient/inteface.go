package jwtclient

import (
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"

	"github.com/filecoin-project/venus-auth/core"

	"github.com/filecoin-project/go-jsonrpc/auth"
)

type IJwtAuthClient interface {
	Verify(ctx context.Context, token string) ([]auth.Permission, error)
}

type jwtAuthClient struct {
	IAuthClient
}

var _ IJwtAuthClient = &jwtAuthClient{}

func (c *jwtAuthClient) Verify(ctx context.Context, token string) ([]auth.Permission, error) {
	res, err := c.IAuthClient.Verify(ctx, token)
	if err != nil {
		return nil, err
	}
	jwtPerms := core.AdaptOldStrategy(res.Perm)
	perms := make([]auth.Permission, len(jwtPerms))
	copy(perms, jwtPerms)
	return perms, nil
}

func WarpIJwtAuthClient(cli IAuthClient) IJwtAuthClient {
	return &jwtAuthClient{IAuthClient: cli}
}

type Logger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
}

func RandSecret() ([]byte, error) {
	return ioutil.ReadAll(io.LimitReader(rand.Reader, 32))
}
