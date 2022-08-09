package jwtclient

import (
	"context"
	"testing"

	venusauth "github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/stretchr/testify/assert"
)

func TestLocalAuthClient(t *testing.T) {
	secret, err := config.RandSecret()
	assert.NoError(t, err)

	payload := venusauth.JWTPayload{
		Perm: core.PermAdmin,
		Name: "MarketLocalToken",
	}

	client, token, err := NewLocalAuthClient(secret, payload)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	permissions, err := client.Verify(ctx, string(token))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 4, len(permissions))
	assert.Contains(t, permissions, core.PermAdmin)
	assert.Contains(t, permissions, core.PermRead)
	assert.Contains(t, permissions, core.PermWrite)
	assert.Contains(t, permissions, core.PermSign)
}
