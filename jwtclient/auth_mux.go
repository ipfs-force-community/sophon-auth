package jwtclient

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/core"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("auth_client")

type opt struct {
	reg *regexp.Regexp
}

type Option func(*opt)

func RegexpOption(reg *regexp.Regexp) Option {
	return func(o *opt) {
		o.reg = reg
	}
}

type trustHandle struct {
	http.Handler
	reg *regexp.Regexp
}

// AuthMux used with jsonrpc library to verify whether the request is legal
type AuthMux struct {
	handler       http.Handler
	local, remote IJwtAuthClient

	trustHandle map[string]trustHandle
}

func NewAuthMux(local, remote IJwtAuthClient, handler http.Handler) *AuthMux {
	return &AuthMux{
		handler:     handler,
		local:       local,
		remote:      remote,
		trustHandle: make(map[string]trustHandle),
	}
}

// TrustHandle for requests that can be accessed directly
// if 'pattern' with '/' as suffix, 'TrustHandler' treat it as a root path,
// that it's all sub-path will be trusted.
// if 'pattern' have with prefix and 'reg' is not nil, use 'reg' check 'pattern'.
// if 'pattern' have no '/' with suffix,
// only the URI exactly matches the 'pattern' would be treat as trusted.
func (authMux *AuthMux) TrustHandle(pattern string, handler http.Handler, opts ...Option) {
	opt := new(opt)
	for _, o := range opts {
		o(opt)
	}
	authMux.trustHandle[pattern] = trustHandle{
		Handler: handler,
		reg:     opt.reg,
	}
}

func (authMux *AuthMux) trustedHandler(uri string) http.Handler {
	// todo: we don't consider the situation that 'trustHandle' is changed in parallelly,
	//  cause we assume trusted handler is static after application initialized
	for trustedURI, handler := range authMux.trustHandle {
		if trustedURI == uri {
			return handler
		}
		if strings.HasPrefix(uri, trustedURI) {
			if trustedURI[len(trustedURI)-1] == '/' {
				return handler
			}
			if handler.reg != nil && handler.reg.MatchString(uri) {
				return handler
			}
		}
	}
	return nil
}

func (authMux *AuthMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := authMux.trustedHandler(r.RequestURI); h != nil {
		h.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	token := r.Header.Get(core.AuthorizationHeader)

	if token == "" {
		token = r.FormValue("token")
		if token != "" {
			token = "Bearer " + token
		}
	}

	if !strings.HasPrefix(token, "Bearer ") {
		log.Warnf("missing Bearer prefix in venusauth header")
		w.WriteHeader(401)
		return
	}

	token = strings.TrimPrefix(token, "Bearer ")

	var err error
	var perm core.Permission
	host := r.RemoteAddr

	ctx = core.CtxWithTokenLocation(ctx, host)

	if !isNil(authMux.local) {
		if perm, err = authMux.local.Verify(ctx, token); err != nil {
			if !isNil(authMux.remote) {
				if perm, err = authMux.remote.Verify(ctx, token); err != nil {
					log.Warnf("JWT Verification failed (originating from %s): %s", r.RemoteAddr, err)
					w.WriteHeader(401)
					return
				}
			} else {
				log.Warnf("JWT Verification failed (originating from %s): %s", r.RemoteAddr, err)
				w.WriteHeader(401)
				return
			}
		}
	} else {
		if !isNil(authMux.remote) {
			if perm, err = authMux.remote.Verify(ctx, token); err != nil {
				log.Warnf("JWT Verification failed (originating from %s): %s", r.RemoteAddr, err)
				w.WriteHeader(401)
				return
			}
		}
	}

	ctx = core.CtxWithPerm(ctx, perm)

	if name, _ := auth.JwtUserFromToken(token); len(name) != 0 {
		ctx = core.CtxWithName(ctx, name)
	}

	*r = *(r.WithContext(ctx))

	authMux.handler.ServeHTTP(w, r)
}

func isNil(ac IJwtAuthClient) bool {
	if ac != nil && !reflect.ValueOf(ac).IsNil() {
		return false
	}
	return true
}
