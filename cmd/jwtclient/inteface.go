package jwtclient

import (
	"context"
	"github.com/filecoin-project/go-jsonrpc/auth"
)

type IJwtAuthAPI interface {
	// Rule[perm:read]
	Verify(ctx context.Context, token string) ([]auth.Permission, error)
	// Rule[perm:read]
	AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error)
}

type IJwtAuthClient interface {
	API() IJwtAuthAPI
	Verify(ctx context.Context, token string) ([]auth.Permission, error)
}

type ILoger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
}
