package auth

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/filecoin-project/venus-auth/log"

	"github.com/filecoin-project/go-address"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/filecoin-project/venus-auth/util"
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

	CreateAccount(ctx context.Context, req *CreateAccountRequest) (*CreateAccountResponse, error)
	UpdateAccount(ctx context.Context, req *UpdateAccountRequest) error
	ListAccounts(ctx context.Context, req *ListAccountsRequest) (ListAccountsResponse, error)
	GetMiner(ctx context.Context, req *GetMinerRequest) (*OutputAccount, error)
	HasMiner(ctx context.Context, req *HasMinerRequest) (bool, error)
	GetAccount(ctx context.Context, req *GetAccountRequest) (*OutputAccount, error)
}

type jwtOAuth struct {
	secret *jwt.HMACSHA
	store  storage.Store
	mp     Mapper
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

	// TODO: remove it next version
	skip, limit := int64(0), int64(20)
	for {
		kps, err := store.List(skip, limit)
		if err != nil {
			return nil, xerrors.Errorf("list token %v", err)
		}
		for _, kp := range kps {
			if len(kp.Secret) == 0 {
				kp.Secret = secret
				log.Infof("update token %s secret %s", kp.Token, secret)
				if err := store.UpdateToken(kp); err != nil {
					return nil, xerrors.Errorf("update token(%s) %v", kp.Token, err)
				}
			}
		}
		if len(kps) == 0 {
			break
		}

		skip += limit
	}

	jwtOAuthInstance = &jwtOAuth{
		secret: jwt.NewHS256(sec),
		store:  store,
		mp:     newMapper(),
	}
	return jwtOAuthInstance, nil
}

func (o *jwtOAuth) GenerateToken(ctx context.Context, pl *JWTPayload) (string, error) {
	// one token, one secret
	secret, err := config.RandSecret()
	if err != nil {
		return "", xerrors.Errorf("rand secret %v", err)
	}
	tk, err := jwt.Sign(pl, jwt.NewHS256(secret))
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

	err = o.store.Put(&storage.KeyPair{Token: token, Secret: hex.EncodeToString(secret), CreateTime: time.Now(), Name: pl.Name, Perm: pl.Perm, Extra: pl.Extra})
	if err != nil {
		return core.EmptyString, xerrors.Errorf("store token failed :%s", err)
	}
	return token.String(), nil
}

func (o *jwtOAuth) Verify(ctx context.Context, token string) (*JWTPayload, error) {
	p := new(JWTPayload)
	tk := []byte(token)

	kp, err := o.store.Get(storage.Token(token))
	if err != nil {
		return nil, xerrors.Errorf("get token: %v", err)
	}
	secret, err := hex.DecodeString(kp.Secret)
	if err != nil {
		return nil, xerrors.Errorf("decode secret %v", err)
	}
	if _, err := jwt.Verify(tk, jwt.NewHS256(secret), p); err != nil {
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

func (o *jwtOAuth) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*CreateAccountResponse, error) {
	exist, err := o.store.HasAccount(req.Name)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("account already exists")
	}
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	mAddr, err := address.NewFromString(req.Miner) // convert address type to local
	if err != nil {
		return nil, err
	}
	newAccount := &storage.Account{
		Id:         uid.String(),
		Name:       req.Name,
		Miner:      mAddr.String(),
		Comment:    req.Comment,
		SourceType: req.SourceType,
		State:      req.State,
		ReqLimit:   req.ReqLimit,
		CreateTime: time.Now().Local(),
		UpdateTime: time.Now().Local(),
	}
	err = o.store.PutAccount(newAccount)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutAccount(newAccount), nil
}

func (o *jwtOAuth) UpdateAccount(ctx context.Context, req *UpdateAccountRequest) error {
	account, err := o.store.GetAccount(req.Name)
	if err != nil {
		return err
	}
	account.UpdateTime = time.Now().Local()
	if req.KeySum&1 == 1 {
		mAddr, err := address.NewFromString(req.Miner)
		if err != nil {
			return err
		}
		account.Miner = mAddr.String()
	}
	if req.KeySum&2 == 2 {
		account.Comment = req.Comment
	}
	if req.KeySum&4 == 4 {
		account.State = req.State
	}
	if req.KeySum&8 == 8 {
		account.SourceType = req.SourceType
	}
	if req.KeySum&16 == 16 {
		account.ReqLimit = req.ReqLimit
	}
	return o.store.UpdateAccount(account)
}

func (o *jwtOAuth) ListAccounts(ctx context.Context, req *ListAccountsRequest) (ListAccountsResponse, error) {
	accounts, err := o.store.ListAccounts(req.GetSkip(), req.GetLimit(), req.State, req.SourceType, req.KeySum)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutAccounts(accounts), nil

}

func (o *jwtOAuth) GetMiner(ctx context.Context, req *GetMinerRequest) (*OutputAccount, error) {
	mAddr, err := address.NewFromString(req.Miner)
	if err != nil {
		return nil, err
	}
	account, err := o.store.GetMiner(mAddr)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutAccount(account), nil
}

func (o *jwtOAuth) HasMiner(ctx context.Context, req *HasMinerRequest) (bool, error) {
	mAddr, err := address.NewFromString(req.Miner)
	if err != nil {
		return false, err
	}
	has, err := o.store.HasMiner(mAddr)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (o *jwtOAuth) GetAccount(ctx context.Context, req *GetAccountRequest) (*OutputAccount, error) {
	account, err := o.store.GetAccount(req.Name)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutAccount(account), nil
}

func DecodeToBytes(enc []byte) ([]byte, error) {
	encoding := base64.RawURLEncoding
	dec := make([]byte, encoding.DecodedLen(len(enc)))
	if _, err := encoding.Decode(dec, enc); err != nil {
		return nil, err
	}
	return dec, nil
}

func JwtAccountFromToken(token string) (string, error) {
	sks := strings.Split(token, ".")
	if len(sks) < 1 {
		return "", fmt.Errorf("can't parse account from input token")

	}
	dec, err := DecodeToBytes([]byte(sks[1]))
	if err != nil {
		return "", err
	}
	payload := &JWTPayload{}
	err = json.Unmarshal(dec, payload)

	return payload.Name, err
}
