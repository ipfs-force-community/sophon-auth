package auth

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/filecoin-project/venus-auth/util"
	"github.com/gbrlsnchs/jwt/v3"
	"golang.org/x/xerrors"
	"time"
)

var (
	ErrorParamsEmpty        = xerrors.New("The mail or password in customParams is empty")
	ErrorRemoveFailed       = xerrors.New("Remove token failed")
	ErrorNonRegisteredToken = xerrors.New("A non-registered token")
	ErrorVerificationFailed = xerrors.New("Verification Failed")
)

var jwtOAuthInstance *jwtOAuth

type VerifyAop interface {
	Verify(ctx context.Context, token string) (*JWTPayload, error)
}

type OAuthService interface {
	GenerateToken(ctx context.Context, cp *JWTPayload) (string, error)
	Verify(ctx context.Context, token string) (*JWTPayload, error)
	RemoveToken(ctx context.Context, token string) error
	Tokens(ctx context.Context, skip, limit int64) ([]*TokenInfo, error)
}

type jwtOAuth struct {
	secret *jwt.HMACSHA
	store  storage.Store
}

type JWTPayload struct {
	Name  string          `json:"name"`
	Perm  core.Permission `json:"perm"`
	Extra string          `json:"ext"`
}

func NewOAuthService(secret string, dbPath string, cnf *config.DBConfig) (OAuthService, error) {
	sec, err := hex.DecodeString(secret)
	if err != nil {
		return nil, err
	}
	store, err := storage.NewStore(cnf, dbPath)
	if err != nil {
		return nil, err
	}
	jwtOAuthInstance = &jwtOAuth{
		secret: jwt.NewHS256(sec),
		store:  store,
	}
	return jwtOAuthInstance, nil
}

func (o *jwtOAuth) GenerateToken(ctx context.Context, pl *JWTPayload) (string, error) {
	tk, err := jwt.Sign(pl, o.secret)
	if err != nil {
		return core.EmptyString, xerrors.Errorf("gen token failed :%s", err)
	}
	token := storage.Token(tk)
	has, err := o.store.Has(token)
	if err != nil {
		return core.EmptyString, err
	}
	if has {
		return token.String(), nil
	}
	err = o.store.Put(&storage.KeyPair{Token: token, CreateTime: time.Now(), Name: pl.Name, Perm: pl.Perm, Extra: pl.Extra})
	if err != nil {
		return core.EmptyString, xerrors.Errorf("store token failed :%s", err)
	}
	return token.String(), nil
}

func (o *jwtOAuth) Verify(ctx context.Context, token string) (*JWTPayload, error) {
	p := new(JWTPayload)
	tk := []byte(token)
	has, err := o.store.Has(storage.Token(token))
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrorNonRegisteredToken
	}
	if _, err := jwt.Verify(tk, o.secret, p); err != nil {
		return nil, ErrorVerificationFailed
	}
	return p, nil
}

type TokenInfo struct {
	Token      string    `json:"token"`
	Name       string    `json:"name"`
	Perm       string    `json:"perm"`
	Custom     string    `json:"custom"`
	CreateTime time.Time `json:"createTime"`
}

func (o *jwtOAuth) Tokens(ctx context.Context, skip, limit int64) ([]*TokenInfo, error) {
	pairs, err := o.store.List(skip, limit)
	if err != nil {
		return nil, err
	}
	tks := make([]*TokenInfo, 0, limit)
	for _, v := range pairs {
		jwtPayload, err := util.JWTPayloadMap(string(v.Token))
		if err != nil {
			return nil, err
		}
		tks = append(tks, &TokenInfo{
			Token:      v.Token.String(),
			CreateTime: v.CreateTime,
			Name:       jwtPayload["name"].(string),
			Perm:       jwtPayload["perm"].(string),
		})
	}
	return tks, nil
}

func (o *jwtOAuth) RemoveToken(ctx context.Context, token string) error {
	tk := []byte(token)
	err := o.store.Delete(storage.Token(tk))
	if err != nil {
		return ErrorRemoveFailed
	}
	return nil
}

func DecodeToBytes(enc []byte) ([]byte, error) {
	encoding := base64.RawURLEncoding
	dec := make([]byte, encoding.DecodedLen(len(enc)))
	if _, err := encoding.Decode(dec, enc); err != nil {
		return nil, err
	}
	return dec, nil
}
