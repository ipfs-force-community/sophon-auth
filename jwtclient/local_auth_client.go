package jwtclient

import (
	"context"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	jwt3 "github.com/gbrlsnchs/jwt/v3"
)

type LocalAuthClient struct {
	alg *jwt3.HMACSHA
}

func NewLocalAuthClient() (*LocalAuthClient, []byte, error) {
	secret, err := config.RandSecret()
	if err != nil {
		return nil, nil, err
	}

	return NewLocalAuthClientWithSecret(secret)
}

func (c *LocalAuthClient) Verify(ctx context.Context, token string) (core.Permission, error) {
	var payload auth.JWTPayload
	_, err := jwt3.Verify([]byte(token), c.alg, &payload)
	if err != nil {
		return "", err
	}

	return payload.Perm, nil
}

func NewLocalAuthClientWithSecret(secret []byte) (*LocalAuthClient, []byte, error) {
	payload := auth.JWTPayload{
		Perm: core.PermAdmin,
		Name: auth.DefaultAdminTokenName,
	}

	client := &LocalAuthClient{
		alg: jwt3.NewHS256(secret),
	}

	token, err := jwt3.Sign(payload, client.alg)
	return client, token, err
}
