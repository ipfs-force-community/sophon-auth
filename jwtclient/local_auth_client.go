package jwtclient

import (
	"context"

	"github.com/filecoin-project/go-jsonrpc/auth"
	venusauth "github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	jwt3 "github.com/gbrlsnchs/jwt/v3"
)

type LocalAuthClient struct {
	alg *jwt3.HMACSHA
}

func NewLocalAuthClient() (*LocalAuthClient, []byte, error) {
	secret, err := RandSecret()
	if err != nil {
		return nil, nil, err
	}

	return NewLocalAuthClientWithSecret(secret)
}

func (c *LocalAuthClient) Verify(ctx context.Context, token string) ([]auth.Permission, error) {
	var payload venusauth.JWTPayload
	_, err := jwt3.Verify([]byte(token), c.alg, &payload)
	if err != nil {
		return nil, err
	}

	jwtPerms := core.AdaptOldStrategy(payload.Perm)
	perms := make([]auth.Permission, len(jwtPerms))
	copy(perms, jwtPerms)
	return perms, nil
}

func NewLocalAuthClientWithSecret(secret []byte) (*LocalAuthClient, []byte, error) {
	payload := venusauth.JWTPayload{
		Perm: core.PermAdmin,
		Name: "defaultLocalToken",
	}

	client := &LocalAuthClient{
		alg: jwt3.NewHS256(secret),
	}

	token, err := jwt3.Sign(payload, client.alg)
	return client, token, err
}
