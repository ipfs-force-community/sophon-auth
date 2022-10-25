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

	"github.com/gin-gonic/gin"

	"github.com/gbrlsnchs/jwt/v3"
	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/filecoin-project/venus-auth/util"
)

var (
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
	RecoverToken(ctx context.Context, token string) error
	Tokens(ctx context.Context, skip, limit int64) ([]*TokenInfo, error)
	GetToken(c context.Context, token string) (*TokenInfo, error)
	GetTokenByName(c context.Context, name string) ([]*TokenInfo, error)

	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
	GetUser(ctx context.Context, req *GetUserRequest) (*OutputUser, error)
	VerifyUsers(ctx context.Context, req *VerifyUsersReq) error
	ListUsers(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, error)
	HasUser(ctx context.Context, req *HasUserRequest) (bool, error)
	UpdateUser(ctx context.Context, req *UpdateUserRequest) error
	DeleteUser(ctx *gin.Context, req *DeleteUserRequest) error
	RecoverUser(ctx *gin.Context, req *RecoverUserRequest) error

	GetUserRateLimits(ctx context.Context, req *GetUserRateLimitsReq) (GetUserRateLimitResponse, error)
	UpsertUserRateLimit(ctx context.Context, req *UpsertUserRateLimitReq) (string, error)
	DelUserRateLimit(ctx context.Context, req *DelUserRateLimitReq) error

	UpsertMiner(ctx context.Context, req *UpsertMinerReq) (bool, error)
	HasMiner(ctx context.Context, req *HasMinerRequest) (bool, error)
	MinerExistInUser(ctx context.Context, req *MinerExistInUserRequest) (bool, error)
	ListMiners(ctx context.Context, req *ListMinerReq) (ListMinerResp, error)
	DelMiner(ctx context.Context, req *DelMinerReq) (bool, error)
	GetUserByMiner(ctx context.Context, req *GetUserByMinerRequest) (*OutputUser, error)

	RegisterSigners(ctx context.Context, req *RegisterSignersReq) error
	SignerExistInUser(ctx context.Context, req *SignerExistInUserReq) (bool, error)
	ListSigner(ctx context.Context, req *ListSignerReq) (ListSignerResp, error)
	UnregisterSigners(ctx context.Context, req *UnregisterSignersReq) error
	HasSigner(ctx context.Context, req *HasSignerReq) (bool, error)
	DelSigner(ctx context.Context, req *DelSignerReq) (bool, error)
	GetUserBySigner(ctx context.Context, req *GetUserBySignerReq) ([]*OutputUser, error)
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

	err = o.store.Put(&storage.KeyPair{
		Token: token, Secret: hex.EncodeToString(secret), CreateTime: time.Now(),
		Name: pl.Name, Perm: pl.Perm, Extra: pl.Extra, IsDeleted: core.NotDelete,
	})
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

func toTokenInfo(kp *storage.KeyPair) (*TokenInfo, error) {
	jwtPayload, err := util.JWTPayloadMap(string(kp.Token))
	if err != nil {
		return nil, err
	}
	return &TokenInfo{
		Token:      kp.Token.String(),
		CreateTime: kp.CreateTime,
		Name:       jwtPayload["name"].(string),
		Perm:       jwtPayload["perm"].(string),
	}, nil
}

func (o *jwtOAuth) GetToken(c context.Context, token string) (*TokenInfo, error) {
	pair, err := o.store.Get(storage.Token(token))
	if err != nil {
		return nil, err
	}
	return toTokenInfo(pair)
}

func (o *jwtOAuth) GetTokenByName(c context.Context, name string) ([]*TokenInfo, error) {
	pairs, err := o.store.ByName(name)
	if err != nil {
		return nil, err
	}
	tokenInfos := make([]*TokenInfo, 0, len(pairs))
	for _, pair := range pairs {
		tokenInfo, err := toTokenInfo(pair)
		if err != nil {
			return nil, err
		}
		tokenInfos = append(tokenInfos, tokenInfo)
	}
	return tokenInfos, nil
}

func (o *jwtOAuth) Tokens(ctx context.Context, skip, limit int64) ([]*TokenInfo, error) {
	pairs, err := o.store.List(skip, limit)
	if err != nil {
		return nil, err
	}
	tks := make([]*TokenInfo, 0, limit)
	for _, pair := range pairs {
		tokenInfo, err := toTokenInfo(pair)
		if err != nil {
			return nil, err
		}
		tks = append(tks, tokenInfo)
	}
	return tks, nil
}

func (o *jwtOAuth) RemoveToken(ctx context.Context, token string) error {
	err := o.store.Delete(storage.Token(token))
	if err != nil {
		return fmt.Errorf("remove token %s: %w", token, err)
	}
	return nil
}

func (o *jwtOAuth) RecoverToken(ctx context.Context, token string) error {
	err := o.store.Recover(storage.Token(token))
	if err != nil {
		return fmt.Errorf("recover token %s: %w", token, err)
	}
	return nil
}

func (o *jwtOAuth) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	exist, err := o.store.HasUser(req.Name)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.New("user already exists")
	}
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	userNew := &storage.User{
		Id:         uid.String(),
		Name:       req.Name,
		State:      req.State,
		CreateTime: time.Now().Local(),
		UpdateTime: time.Now().Local(),
		IsDeleted:  core.NotDelete,
	}
	if req.Comment != nil {
		userNew.Comment = *req.Comment
	}
	err = o.store.PutUser(userNew)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(userNew), nil
}

func (o *jwtOAuth) UpdateUser(ctx context.Context, req *UpdateUserRequest) error {
	user, err := o.store.GetUser(req.Name)
	if err != nil {
		return err
	}
	user.UpdateTime = time.Now().Local()
	if req.Comment != nil {
		user.Comment = *req.Comment
	}
	if req.State != core.UserStateUndefined {
		user.State = req.State
	}
	return o.store.UpdateUser(user)
}

func (o *jwtOAuth) VerifyUsers(ctx context.Context, req *VerifyUsersReq) error {
	return o.store.VerifyUsers(req.Names)
}

func (o *jwtOAuth) ListUsers(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, error) {
	users, err := o.store.ListUsers(req.GetSkip(), req.GetLimit(), core.UserState(req.State))
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUsers(users), nil
}

func (o *jwtOAuth) HasUser(ctx context.Context, req *HasUserRequest) (bool, error) {
	return o.store.HasUser(req.Name)
}

func (o *jwtOAuth) DeleteUser(ctx *gin.Context, req *DeleteUserRequest) error {
	return o.store.DeleteUser(req.Name)
}

func (o *jwtOAuth) RecoverUser(ctx *gin.Context, req *RecoverUserRequest) error {
	return o.store.RecoverUser(req.Name)
}

func (o *jwtOAuth) GetUserByMiner(ctx context.Context, req *GetUserByMinerRequest) (*OutputUser, error) {
	mAddr, err := address.NewFromString(req.Miner)
	if err != nil {
		return nil, err
	}
	user, err := o.store.GetUserByMiner(mAddr)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(user), nil
}

func (o *jwtOAuth) GetUserBySigner(ctx context.Context, req *GetUserBySignerReq) ([]*OutputUser, error) {
	addr, err := address.NewFromString(req.Signer)
	if err != nil {
		return nil, err
	}
	users, err := o.store.GetUserBySigner(addr)
	if err != nil {
		return nil, err
	}

	outUsers := make([]*OutputUser, len(users))
	for idx, user := range users {
		outUsers[idx] = o.mp.ToOutPutUser(user)
	}

	return outUsers, nil
}

func (o *jwtOAuth) GetUser(ctx context.Context, req *GetUserRequest) (*OutputUser, error) {
	user, err := o.store.GetUser(req.Name)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(user), nil
}

func (o jwtOAuth) GetUserRateLimits(ctx context.Context, req *GetUserRateLimitsReq) (GetUserRateLimitResponse, error) {
	return o.store.GetRateLimits(req.Name, req.Id)
}

func (o *jwtOAuth) UpsertUserRateLimit(ctx context.Context, req *UpsertUserRateLimitReq) (string, error) {
	return o.store.PutRateLimit((*storage.UserRateLimit)(req))
}

func (o jwtOAuth) DelUserRateLimit(ctx context.Context, req *DelUserRateLimitReq) error {
	return o.store.DelRateLimit(req.Name, req.Id)
}

func (o *jwtOAuth) UpsertMiner(ctx context.Context, req *UpsertMinerReq) (bool, error) {
	maddr, err := address.NewFromString(req.Miner)
	if err != nil || maddr.Empty() {
		return false, xerrors.Errorf("invalid miner address:%s, error: %w", req.Miner, err)
	}

	if maddr.Protocol() != address.ID {
		return false, fmt.Errorf("invalid protocol type: %v", maddr.Protocol())
	}

	return o.store.UpsertMiner(maddr, req.User, req.OpenMining)
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

func (o *jwtOAuth) MinerExistInUser(ctx context.Context, req *MinerExistInUserRequest) (bool, error) {
	mAddr, err := address.NewFromString(req.Miner)
	if err != nil {
		return false, err
	}

	exist, err := o.store.MinerExistInUser(mAddr, req.User)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (o *jwtOAuth) ListMiners(ctx context.Context, req *ListMinerReq) (ListMinerResp, error) {
	miners, err := o.store.ListMiners(req.User)
	if err != nil {
		return nil, xerrors.Errorf("list user:%s miners failed:%w", req.User, err)
	}

	outs := make([]*OutputMiner, len(miners))
	for idx, m := range miners {
		addrStr := m.Miner.Address().String()
		outs[idx] = &OutputMiner{
			Miner:      addrStr,
			User:       m.User,
			OpenMining: *m.OpenMining,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		}
	}
	return outs, nil
}

func (o jwtOAuth) DelMiner(ctx context.Context, req *DelMinerReq) (bool, error) {
	miner, err := address.NewFromString(req.Miner)
	if err != nil {
		return false, xerrors.Errorf("invalid miner address:%s, %w", req.Miner, err)
	}
	return o.store.DelMiner(miner)
}

func (o *jwtOAuth) RegisterSigners(ctx context.Context, req *RegisterSignersReq) error {
	for _, signer := range req.Signers {
		addr, err := address.NewFromString(signer)
		if err != nil || addr.Empty() {
			return fmt.Errorf("invalid signer address: %s, error: %w", signer, err)
		}

		if !isSignerAddress(addr) {
			return fmt.Errorf("invalid protocol type: %v", addr.Protocol())
		}

		err = o.store.RegisterSigner(addr, req.User)
		if err != nil {
			return fmt.Errorf("unregister signer:%s, error: %w", signer, err)
		}
	}

	return nil
}

func (o *jwtOAuth) SignerExistInUser(ctx context.Context, req *SignerExistInUserReq) (bool, error) {
	addr, err := address.NewFromString(req.Signer)
	if err != nil {
		return false, err
	}

	if !isSignerAddress(addr) {
		return false, fmt.Errorf("invalid protocol type: %v", addr.Protocol())
	}

	has, err := o.store.SignerExistInUser(addr, req.User)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (o *jwtOAuth) ListSigner(ctx context.Context, req *ListSignerReq) (ListSignerResp, error) {
	signers, err := o.store.ListSigner(req.User)
	if err != nil {
		return nil, xerrors.Errorf("list user:%s signer failed: %w", req.User, err)
	}

	outs := make([]*OutputSigner, len(signers))
	for idx, m := range signers {
		addrStr := m.Signer.Address().String()
		outs[idx] = &OutputSigner{
			Signer:    addrStr,
			User:      m.User,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}
	return outs, nil
}

func (o *jwtOAuth) UnregisterSigners(ctx context.Context, req *UnregisterSignersReq) error {
	for _, signer := range req.Signers {
		addr, err := address.NewFromString(signer)
		if err != nil || addr.Empty() {
			return fmt.Errorf("invalid signer address: %s, error: %w", signer, err)
		}

		if !isSignerAddress(addr) {
			return fmt.Errorf("invalid protocol type: %v", addr.Protocol())
		}

		err = o.store.UnregisterSigner(addr, req.User)
		if err != nil {
			return fmt.Errorf("unregister signer:%s, error: %w", signer, err)
		}
	}

	return nil
}

func (o jwtOAuth) HasSigner(ctx context.Context, req *HasSignerReq) (bool, error) {
	addr, err := address.NewFromString(req.Signer)
	if err != nil {
		return false, xerrors.Errorf("invalid signer address:%s, %w", req.Signer, err)
	}

	if !isSignerAddress(addr) {
		return false, fmt.Errorf("invalid protocol type: %v", addr.Protocol())
	}

	return o.store.HasSigner(addr)
}

func (o jwtOAuth) DelSigner(ctx context.Context, req *DelSignerReq) (bool, error) {
	addr, err := address.NewFromString(req.Signer)
	if err != nil {
		return false, xerrors.Errorf("invalid signer address:%s, %w", req.Signer, err)
	}

	if !isSignerAddress(addr) {
		return false, fmt.Errorf("invalid protocol type: %v", addr.Protocol())
	}

	return o.store.DelSigner(addr)
}

func DecodeToBytes(enc []byte) ([]byte, error) {
	encoding := base64.RawURLEncoding
	dec := make([]byte, encoding.DecodedLen(len(enc)))
	if _, err := encoding.Decode(dec, enc); err != nil {
		return nil, err
	}
	return dec, nil
}

func JwtUserFromToken(token string) (string, error) {
	sks := strings.Split(token, ".")
	if len(sks) < 1 {
		return "", fmt.Errorf("can't parse user from input token")
	}
	dec, err := DecodeToBytes([]byte(sks[1]))
	if err != nil {
		return "", err
	}
	payload := &JWTPayload{}
	err = json.Unmarshal(dec, payload)

	return payload.Name, err
}

func isSignerAddress(addr address.Address) bool {
	return addr.Protocol() == address.SECP256K1 || addr.Protocol() == address.BLS
}
