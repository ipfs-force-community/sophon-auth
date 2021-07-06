package jwtclient

import (
	"context"
	"fmt"
	"github.com/filecoin-project/go-jsonrpc/auth"
	auth2 "github.com/filecoin-project/venus-auth/auth"
	ipfsHttp "github.com/ipfs/go-ipfs-cmds/http"
	"net/http"
	"strings"
)

type CtxKey int

const (
	accountKey CtxKey = iota
	tokenLocationkey
)

// AuthMux used with jsonrpc library to verify whether the request is legal
type AuthMux struct {
	ILoger
	handler       http.Handler
	local, remote IJwtAuthClient

	trustHandle map[string]http.Handler
}

func NewAuthMux(local, remote IJwtAuthClient, handler http.Handler, loger ILoger) *AuthMux {
	return &AuthMux{handler: handler,
		local: local, remote: remote,
		trustHandle: make(map[string]http.Handler), ILoger: loger}
}

// TrustHandle for requests that can be accessed directly
func (authMux *AuthMux) TrustHandle(pattern string, handler http.Handler) {
	authMux.trustHandle[pattern] = handler
}

func (authMux *AuthMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handle, ok := authMux.trustHandle[r.RequestURI]; ok {
		handle.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	token := r.Header.Get("Authorization")

	if token == "" {
		token = r.FormValue("token")
		if token != "" {
			token = "Bearer " + token
		}
	}

	if !strings.HasPrefix(token, "Bearer ") {
		authMux.Warnf("missing Bearer prefix in venusauth header")
		w.WriteHeader(401)
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")

	var perms []auth.Permission
	var err error
	var tokenLocation = "local"

	if perms, err = authMux.local.Verify(ctx, token); err != nil {
		if perms, err = authMux.remote.Verify(ctx, token); err != nil {
			authMux.Warnf("JWT Verification failed (originating from %s): %s", r.RemoteAddr, err)
			w.WriteHeader(401)
			return
		}
		tokenLocation = "remote"
	}

	ctx = auth.WithPerm(ctx, perms)
	ctx = ipfsHttp.WithPerm(ctx, perms)
	ctx = CtxWithTokenLocation(ctx, tokenLocation)

	if name, _ := auth2.JwtUserFromToken(token); len(name) != 0 {
		ctx = CtxWithName(ctx, name)
	}

	*r = *(r.WithContext(ctx))

	authMux.handler.ServeHTTP(w, r)
}

func (authMux *AuthMux) Warnf(template string, args ...interface{}) {
	if authMux.ILoger == nil {
		fmt.Printf("auth-middware warning:%s\n", fmt.Sprintf(template, args...))
		return
	}
	authMux.ILoger.Warnf(template, args...)
}

func (authMux *AuthMux) Infof(template string, args ...interface{}) {
	if authMux.ILoger == nil {
		fmt.Printf("auth-midware info:%s\n", fmt.Sprintf(template, args...))
		return
	}
	authMux.ILoger.Infof(template, args...)
}

func (authMux *AuthMux) Errorf(template string, args ...interface{}) {
	if authMux.ILoger == nil {
		fmt.Printf("auth-midware error:%s\n", fmt.Sprintf(template, args...))
		return
	}
	authMux.ILoger.Errorf(template, args...)
}

func ctxWithString(ctx context.Context, k CtxKey, v string) context.Context {
	return context.WithValue(ctx, k, v)
}

func ctxGetString(ctx context.Context, k CtxKey) (v string, exists bool) {
	iv := ctx.Value(k)
	if iv != nil {
		return
	}
	v, exists = iv.(string)
	return
}

func CtxWithName(ctx context.Context, v string) context.Context {
	return ctxWithString(ctx, accountKey, v)
}

func CtxGetName(ctx context.Context) (name string, exists bool) {
	return ctxGetString(ctx, accountKey)
}

func CtxWithTokenLocation(ctx context.Context, v string) context.Context {
	return ctxWithString(ctx, tokenLocationkey, v)
}

func CtxGetTokenLocation(ctx context.Context) (location string, exists bool) {
	return ctxGetString(ctx, tokenLocationkey)
}
