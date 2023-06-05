package core

import (
	"context"
)

type CtxKey int

const (
	accountKey CtxKey = iota
	tokenLocationKey
	permKey
)

func HasPerm(ctx context.Context, defaultPerms []Permission, perm Permission) bool {
	callerPerms, ok := CtxGetPerm(ctx)
	if !ok {
		callerPerms = defaultPerms
	}

	for _, callerPerm := range callerPerms {
		if callerPerm == perm {
			return true
		}
	}
	return false
}

func CtxWithPerm(ctx context.Context, perm Permission) context.Context {
	return context.WithValue(ctx, permKey, AdaptOldStrategy(perm))
}

func CtxWithPerms(ctx context.Context, perms []Permission) context.Context {
	return context.WithValue(ctx, permKey, perms)
}

func CtxGetPerm(ctx context.Context) ([]Permission, bool) {
	v, exist := ctx.Value(permKey).([]Permission)

	return v, exist
}

func ctxGetString(ctx context.Context, k CtxKey) (v string, exists bool) {
	v, exists = ctx.Value(k).(string)
	return
}

func CtxWithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, accountKey, name)
}

func CtxGetName(ctx context.Context) (name string, exists bool) {
	return ctxGetString(ctx, accountKey)
}

func CtxWithTokenLocation(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenLocationKey, token)
}

func CtxGetTokenLocation(ctx context.Context) (location string, exists bool) {
	return ctxGetString(ctx, tokenLocationKey)
}

type ValueFromCtx struct{}

func (vfc *ValueFromCtx) AccFromCtx(ctx context.Context) (string, bool) {
	return CtxGetName(ctx)
}

func (vfc *ValueFromCtx) HostFromCtx(ctx context.Context) (string, bool) {
	return CtxGetTokenLocation(ctx)
}
