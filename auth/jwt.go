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
	ErrorPermissionDenied   = xerrors.New("Permission Deny")
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
	DeleteUser(ctx context.Context, req *DeleteUserRequest) error
	RecoverUser(ctx context.Context, req *RecoverUserRequest) error

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
	if !isAdmin(ctx) {
		return "", ErrorPermissionDenied
	}

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
	if !hasPerm(ctx, core.PermRead) {
		return nil, ErrorPermissionDenied
	}

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
	if !isAdmin(c) {
		return nil, ErrorPermissionDenied
	}

	pair, err := o.store.Get(storage.Token(token))
	if err != nil {
		return nil, err
	}
	return toTokenInfo(pair)
}

func (o *jwtOAuth) GetTokenByName(c context.Context, name string) ([]*TokenInfo, error) {
	if !isAdmin(c) && !isUserOwner(c, name) {
		return nil, ErrorPermissionDenied
	}

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
	if !isAdmin(ctx) {
		return nil, ErrorPermissionDenied
	}

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
	if !isAdmin(ctx) && !isTokenOwner(ctx, token) {
		return ErrorPermissionDenied
	}

	err := o.store.Delete(storage.Token(token))
	if err != nil {
		return fmt.Errorf("remove token %s: %w", token, err)
	}
	return nil
}

func (o *jwtOAuth) RecoverToken(ctx context.Context, token string) error {
	if !isAdmin(ctx) && !isTokenOwner(ctx, token) {
		return ErrorPermissionDenied
	}

	err := o.store.Recover(storage.Token(token))
	if err != nil {
		return fmt.Errorf("recover token %s: %w", token, err)
	}
	return nil
}

func (o *jwtOAuth) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	if !isAdmin(ctx) {
		return nil, ErrorPermissionDenied
	}

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
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}

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
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}
	return o.store.VerifyUsers(req.Names)
}

func (o *jwtOAuth) ListUsers(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, error) {
	if !isAdmin(ctx) {
		return nil, ErrorPermissionDenied
	}

	users, err := o.store.ListUsers(req.GetSkip(), req.GetLimit(), core.UserState(req.State))
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUsers(users), nil
}

func (o *jwtOAuth) HasUser(ctx context.Context, req *HasUserRequest) (bool, error) {
	if !isAdmin(ctx) {
		return false, ErrorPermissionDenied
	}

	return o.store.HasUser(req.Name)
}

func (o *jwtOAuth) DeleteUser(ctx context.Context, req *DeleteUserRequest) error {
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}
	return o.store.DeleteUser(req.Name)
}

func (o *jwtOAuth) RecoverUser(ctx context.Context, req *RecoverUserRequest) error {
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}

	return o.store.RecoverUser(req.Name)
}

func (o *jwtOAuth) GetUserByMiner(ctx context.Context, req *GetUserByMinerRequest) (*OutputUser, error) {
	if !isAdmin(ctx) {
		return nil, ErrorPermissionDenied
	}

	user, err := o.store.GetUserByMiner(req.Miner)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(user), nil
}

func (o *jwtOAuth) GetUserBySigner(ctx context.Context, req *GetUserBySignerReq) ([]*OutputUser, error) {
	if !isAdmin(ctx) {
		return nil, ErrorPermissionDenied
	}

	users, err := o.store.GetUserBySigner(req.Signer)
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
	if !isAdmin(ctx) && !isUserOwner(ctx, req.Name) {
		return nil, ErrorPermissionDenied
	}
	user, err := o.store.GetUser(req.Name)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(user), nil
}

func (o jwtOAuth) GetUserRateLimits(ctx context.Context, req *GetUserRateLimitsReq) (GetUserRateLimitResponse, error) {
	if !isAdmin(ctx) {
		return nil, ErrorPermissionDenied
	}

	return o.store.GetRateLimits(req.Name, req.Id)
}

func (o *jwtOAuth) UpsertUserRateLimit(ctx context.Context, req *UpsertUserRateLimitReq) (string, error) {
	if !isAdmin(ctx) {
		return "", ErrorPermissionDenied
	}

	return o.store.PutRateLimit((*storage.UserRateLimit)(req))
}

func (o jwtOAuth) DelUserRateLimit(ctx context.Context, req *DelUserRateLimitReq) error {
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}

	return o.store.DelRateLimit(req.Name, req.Id)
}

func (o *jwtOAuth) UpsertMiner(ctx context.Context, req *UpsertMinerReq) (bool, error) {
	if !isAdmin(ctx) {
		return false, ErrorPermissionDenied
	}

	mAddr := req.Miner
	if mAddr.Protocol() != address.ID {
		return false, fmt.Errorf("invalid protocol type: %v", mAddr.Protocol())
	}

	return o.store.UpsertMiner(mAddr, req.User, req.OpenMining)
}

func (o *jwtOAuth) HasMiner(ctx context.Context, req *HasMinerRequest) (bool, error) {
	if !isAdmin(ctx) {
		return false, ErrorPermissionDenied
	}

	has, err := o.store.HasMiner(req.Miner)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (o *jwtOAuth) MinerExistInUser(ctx context.Context, req *MinerExistInUserRequest) (bool, error) {
	if !isAdmin(ctx) && !isUserOwner(ctx, req.User) {
		return false, ErrorPermissionDenied
	}

	exist, err := o.store.MinerExistInUser(req.Miner, req.User)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (o *jwtOAuth) ListMiners(ctx context.Context, req *ListMinerReq) (ListMinerResp, error) {
	if !isAdmin(ctx) && !isUserOwner(ctx, req.User) {
		return nil, ErrorPermissionDenied
	}

	miners, err := o.store.ListMiners(req.User)
	if err != nil {
		return nil, xerrors.Errorf("list user:%s miners failed:%w", req.User, err)
	}

	outs := make([]*OutputMiner, len(miners))
	for idx, m := range miners {
		outs[idx] = &OutputMiner{
			Miner:      m.Miner.Address(),
			User:       m.User,
			OpenMining: *m.OpenMining,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
		}
	}
	return outs, nil
}

func (o jwtOAuth) DelMiner(ctx context.Context, req *DelMinerReq) (bool, error) {

	if !isAdmin(ctx) && !isMinerOwner(ctx, o.store, req.Miner) {
		return false, ErrorPermissionDenied
	}

	return o.store.DelMiner(req.Miner)
}

func (o *jwtOAuth) RegisterSigners(ctx context.Context, req *RegisterSignersReq) error {
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}

	for _, signer := range req.Signers {
		if !isSignerAddress(signer) {
			return fmt.Errorf("invalid protocol type: %v", signer.Protocol())
		}

		err := o.store.RegisterSigner(signer, req.User)
		if err != nil {
			return fmt.Errorf("unregister signer:%s, error: %w", signer, err)
		}
	}

	return nil
}

func (o *jwtOAuth) SignerExistInUser(ctx context.Context, req *SignerExistInUserReq) (bool, error) {
	if !isAdmin(ctx) && !isUserOwner(ctx, req.User) {
		return false, ErrorPermissionDenied
	}

	addr := req.Signer
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
	if !isAdmin(ctx) && !isUserOwner(ctx, req.User) {
		return nil, ErrorPermissionDenied
	}

	signers, err := o.store.ListSigner(req.User)
	if err != nil {
		return nil, xerrors.Errorf("list user:%s signer failed: %w", req.User, err)
	}

	outs := make([]*OutputSigner, len(signers))
	for idx, m := range signers {
		outs[idx] = &OutputSigner{
			Signer:    m.Signer.Address(),
			User:      m.User,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
	}
	return outs, nil
}

func (o *jwtOAuth) UnregisterSigners(ctx context.Context, req *UnregisterSignersReq) error {
	if !isAdmin(ctx) {
		return ErrorPermissionDenied
	}

	for _, signer := range req.Signers {
		if !isSignerAddress(signer) {
			return fmt.Errorf("invalid protocol type: %v", signer.Protocol())
		}

		err := o.store.UnregisterSigner(signer, req.User)
		if err != nil {
			return fmt.Errorf("unregister signer:%s, error: %w", signer, err)
		}
	}

	return nil
}

func (o jwtOAuth) HasSigner(ctx context.Context, req *HasSignerReq) (bool, error) {
	if !isAdmin(ctx) {
		return false, ErrorPermissionDenied
	}

	addr := req.Signer
	if !isSignerAddress(addr) {
		return false, fmt.Errorf("invalid protocol type: %v", addr.Protocol())
	}

	return o.store.HasSigner(addr)
}

func (o jwtOAuth) DelSigner(ctx context.Context, req *DelSignerReq) (bool, error) {
	addr := req.Signer
	if !isAdmin(ctx) && !isSignerOwner(ctx, o.store, addr) {
		return false, ErrorPermissionDenied
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

func isAdmin(ctx context.Context) bool {
	callerPerms := ctxGetPerm(ctx)
	for _, callerPerm := range callerPerms {
		if callerPerm == core.PermAdmin {
			return true
		}
	}
	return false
}

func hasPerm(ctx context.Context, perm core.Permission) bool {
	callerPerms := ctxGetPerm(ctx)
	for _, callerPerm := range callerPerms {
		if callerPerm == perm {
			return true
		}
	}
	return false
}

func isTokenOwner(ctx context.Context, token string) bool {
	userName := ctxGetName(ctx)
	tokenName, err := JwtUserFromToken(token)
	if err != nil {
		return false
	}
	return userName == tokenName
}

func isUserOwner(ctx context.Context, user string) bool {
	userName := ctxGetName(ctx)
	return userName == user
}

type minerOwnershipChecker interface {
	MinerExistInUser(maddr address.Address, userName string) (bool, error)
}

func isMinerOwner(ctx context.Context, checker minerOwnershipChecker, miner address.Address) bool {
	userName := ctxGetName(ctx)
	has, err := checker.MinerExistInUser(miner, userName)
	if err != nil {
		return false
	}
	return has
}

type signerOwnershipChecker interface {
	SignerExistInUser(signer address.Address, userName string) (bool, error)
}

func isSignerOwner(ctx context.Context, checker signerOwnershipChecker, signer address.Address) bool {
	userName := ctxGetName(ctx)
	has, err := checker.SignerExistInUser(signer, userName)
	if err != nil {
		return false
	}
	return has
}
