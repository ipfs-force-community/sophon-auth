package jwtclient

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc/auth"
)

type IJwtAuthClient interface {
	Verify(ctx context.Context, token string) ([]auth.Permission, error)
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
