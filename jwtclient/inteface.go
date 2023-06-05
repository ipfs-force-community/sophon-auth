package jwtclient

import (
	"context"

	"github.com/ipfs-force-community/sophon-auth/core"
)

type IJwtAuthClient interface {
	Verify(ctx context.Context, token string) (core.Permission, error)
}

type jwtAuthClient struct {
	IAuthClient
}

var _ IJwtAuthClient = &jwtAuthClient{}

func (c *jwtAuthClient) Verify(ctx context.Context, token string) (core.Permission, error) {
	res, err := c.IAuthClient.Verify(ctx, token)
	if err != nil {
		return "", err
	}

	return res.Perm, nil
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
