package auth

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/ipfs-force-community/venus-auth/core"
	"github.com/ipfs-force-community/venus-auth/db"
	"github.com/ipfs-force-community/venus-auth/util"
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
	Tokens(ctx context.Context, pageIndex, pageSize int64) ([]*TokenInfo, error)
	// limit token generate,use to protect the service when a malicious user registers a new token
	// true: lock    false: unlock
	//KeepCalm(bool) error
}

type jwtOAuth struct {
	secret *jwt.HMACSHA
	store  db.Database
}

type JWTPayload struct {
	Name  string          `json:"name"`
	Perm  core.Permission `json:"perm"`
	Extra string          `json:"ext"`
}

func NewOAuthService(secret string, dbPath string) (OAuthService, error) {
	sec, err := hex.DecodeString(secret)
	if err != nil {
		return nil, err
	}
	store, err := db.Open(dbPath)
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
	token := string(tk)
	val, err := time.Now().MarshalBinary()
	if err != nil {
		return core.EmptyString, xerrors.Errorf("failed to marshal time :%s", err)
	}
	err = o.store.Put(tk, val)
	if err != nil {
		return core.EmptyString, xerrors.Errorf("store token failed :%s", err)
	}
	return token, nil
}

func (o *jwtOAuth) Verify(ctx context.Context, token string) (*JWTPayload, error) {
	p := new(JWTPayload)
	tk := []byte(token)
	_, err := o.store.Get(tk)
	if err != nil {
		if err.Error() == "Key not found" {
			return nil, ErrorNonRegisteredToken
		}
		return nil, err
	}
	if _, err := jwt.Verify(tk, o.secret, p); err != nil {
		return nil, ErrorVerificationFailed
	}

	return p, nil
}

type TokenInfo struct {
	Token     string    `json:"token"`
	Name      string    `json:"name"`
	CreatTime time.Time `json:"createTime"`
}

func (o *jwtOAuth) Tokens(ctx context.Context, pageIndex, pageSize int64) ([]*TokenInfo, error) {
	skip := (pageIndex - 1) * pageSize
	pairs, err := o.store.Fetch(skip, pageSize)
	if err != nil {
		return nil, err
	}
	tks := make([]*TokenInfo, 0, pageSize)
	for ch := range pairs {
		tm := time.Time{}
		err = tm.UnmarshalBinary(ch.Val)
		if err != nil {
			return nil, err
		}
		jwtPayload, err := util.JWTPayloadMap(string(ch.Key))
		if err != nil {
			return nil, err
		}
		tks = append(tks, &TokenInfo{
			Token:     string(ch.Key),
			CreatTime: tm,
			Name:      jwtPayload["name"].(string),
		})
	}
	return tks, nil
}
func (o *jwtOAuth) RemoveToken(ctx context.Context, token string) error {
	tk := []byte(token)
	err := o.store.Remove(tk)
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
