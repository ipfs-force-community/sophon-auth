package jwtclient

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/sophon-auth/core"
)

type mockImp struct{}

func (m mockImp) Verify(ctx context.Context, token string) (core.Permission, error) {
	panic("implement me")
}

var _ IJwtAuthClient = (*mockImp)(nil)

type authCli struct {
	c IJwtAuthClient
}

func TestIsNil(t *testing.T) {
	ac := authCli{}
	assert.True(t, isNil(ac.c))

	var imp *mockImp
	ac2 := authCli{c: imp}
	assert.True(t, isNil(ac2.c))

	ac3 := authCli{c: &mockImp{}}
	assert.False(t, isNil(ac3.c))
}

func TestTruthHandle(t *testing.T) {
	reg := regexp.MustCompile(`/piece/[a-z0-9]*`)
	mux := NewAuthMux(nil, nil, nil)
	mock := &http.ServeMux{}
	mux.TrustHandle("/piece/", mock, RegexpOption(reg))
	path := "/piece/xfdfs1fs"
	assert.NotNil(t, mux.trustedHandler(path))
}
