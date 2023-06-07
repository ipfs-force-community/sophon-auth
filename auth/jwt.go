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

	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/storage"
	"github.com/ipfs-force-community/sophon-auth/util"
)

var (
	ErrorNonRegisteredToken = xerrors.New("A non-registered token")
	ErrorVerificationFailed = xerrors.New("Verification Failed")
	ErrorPermissionDeny     = xerrors.New("Permission Deny")
	ErrorPermissionNotFound = errors.New("permission not found")
	ErrorUsernameNotFound   = errors.New("username not found")
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
	store storage.Store
	mp    Mapper
}

type JWTPayload struct {
	Name  string          `json:"name"`
	Perm  core.Permission `json:"perm"`
	Extra string          `json:"ext"`
}

func NewOAuthService(dbPath string, cnf *config.DBConfig) (OAuthService, error) {
	store, err := storage.NewStore(cnf, dbPath)
	if err != nil {
		return nil, err
	}

	jwtOAuthInstance = &jwtOAuth{
		store: store,
		mp:    newMapper(),
	}
	return jwtOAuthInstance, nil
}

func (o *jwtOAuth) GenerateToken(ctx context.Context, pl *JWTPayload) (string, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return "", fmt.Errorf("need admin prem: %w", err)
	}

	exist, err := o.store.HasUser(pl.Name)
	if err != nil {
		return "", fmt.Errorf("check user %s exist failed: %w", pl.Name, err)
	}
	if !exist {
		return "", fmt.Errorf("token must be based on an existing user %s to generate", pl.Name)
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
	err := permCheck(ctx, core.PermRead)
	if err != nil {
		return nil, fmt.Errorf("need read prem: %w", err)
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

func (o *jwtOAuth) GetToken(ctx context.Context, token string) (*TokenInfo, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
	}

	pair, err := o.store.Get(storage.Token(token))
	if err != nil {
		return nil, err
	}
	return toTokenInfo(pair)
}

func (o *jwtOAuth) GetTokenByName(ctx context.Context, username string) ([]*TokenInfo, error) {
	err := userPermCheck(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("need admin prem or user %s does not match: %w", username, err)
	}

	pairs, err := o.store.ByName(username)
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
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
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
	err := tokenPermCheck(ctx, token)
	if err != nil {
		return fmt.Errorf("need admin prem or token %s check failed: %w", token, err)
	}

	err = o.store.Delete(storage.Token(token))
	if err != nil {
		return fmt.Errorf("remove token %s: %w", token, err)
	}
	return nil
}

func (o *jwtOAuth) RecoverToken(ctx context.Context, token string) error {
	err := tokenPermCheck(ctx, token)
	if err != nil {
		return fmt.Errorf("need admin prem or token %s check failed: %w", token, err)
	}

	err = o.store.Recover(storage.Token(token))
	if err != nil {
		return fmt.Errorf("recover token %s: %w", token, err)
	}
	return nil
}

func (o *jwtOAuth) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
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
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
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
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
	}

	return o.store.VerifyUsers(req.Names)
}

func (o *jwtOAuth) ListUsers(ctx context.Context, req *ListUsersRequest) (ListUsersResponse, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
	}

	users, err := o.store.ListUsers(req.GetSkip(), req.GetLimit(), core.UserState(req.State))
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUsers(users), nil
}

func (o *jwtOAuth) HasUser(ctx context.Context, req *HasUserRequest) (bool, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return false, fmt.Errorf("need admin prem: %w", err)
	}

	return o.store.HasUser(req.Name)
}

func (o *jwtOAuth) DeleteUser(ctx context.Context, req *DeleteUserRequest) error {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
	}

	return o.store.DeleteUser(req.Name)
}

func (o *jwtOAuth) RecoverUser(ctx context.Context, req *RecoverUserRequest) error {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
	}
	return o.store.RecoverUser(req.Name)
}

func (o *jwtOAuth) GetUserByMiner(ctx context.Context, req *GetUserByMinerRequest) (*OutputUser, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
	}

	user, err := o.store.GetUserByMiner(req.Miner)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(user), nil
}

func (o *jwtOAuth) GetUserBySigner(ctx context.Context, req *GetUserBySignerReq) ([]*OutputUser, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
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
	err := userPermCheck(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("need admin prem or user %s does not match: %w", req.Name, err)
	}

	user, err := o.store.GetUser(req.Name)
	if err != nil {
		return nil, err
	}
	return o.mp.ToOutPutUser(user), nil
}

func (o jwtOAuth) GetUserRateLimits(ctx context.Context, req *GetUserRateLimitsReq) (GetUserRateLimitResponse, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return nil, fmt.Errorf("need admin prem: %w", err)
	}

	return o.store.GetRateLimits(req.Name, req.Id)
}

func (o *jwtOAuth) UpsertUserRateLimit(ctx context.Context, req *UpsertUserRateLimitReq) (string, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return "nil", fmt.Errorf("need admin prem: %w", err)
	}

	return o.store.PutRateLimit((*storage.UserRateLimit)(req))
}

func (o jwtOAuth) DelUserRateLimit(ctx context.Context, req *DelUserRateLimitReq) error {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
	}

	return o.store.DelRateLimit(req.Name, req.Id)
}

func (o *jwtOAuth) UpsertMiner(ctx context.Context, req *UpsertMinerReq) (bool, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return false, fmt.Errorf("need admin prem: %w", err)
	}

	mAddr := req.Miner
	if mAddr.Protocol() != address.ID {
		return false, fmt.Errorf("invalid protocol type: %v", mAddr.Protocol())
	}

	return o.store.UpsertMiner(mAddr, req.User, req.OpenMining)
}

func (o *jwtOAuth) HasMiner(ctx context.Context, req *HasMinerRequest) (bool, error) {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return false, fmt.Errorf("need admin prem: %w", err)
	}

	has, err := o.store.HasMiner(req.Miner)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (o *jwtOAuth) MinerExistInUser(ctx context.Context, req *MinerExistInUserRequest) (bool, error) {
	err := userPermCheck(ctx, req.User)
	if err != nil {
		return false, fmt.Errorf("need admin prem or user %s does not match: %w", req.User, err)
	}

	exist, err := o.store.MinerExistInUser(req.Miner, req.User)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (o *jwtOAuth) ListMiners(ctx context.Context, req *ListMinerReq) (ListMinerResp, error) {
	err := userPermCheck(ctx, req.User)
	if err != nil {
		return nil, fmt.Errorf("need admin prem or user %s does not match: %w", req.User, err)
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
	if permCheck(ctx, core.PermAdmin) != nil {
		if err := ownerOfMinerCheck(ctx, o.store, req.Miner); err != nil {
			return false, fmt.Errorf("need admin prem or %s ownership check error: %w", req.Miner, err)
		}
	}

	return o.store.DelMiner(req.Miner)
}

func (o *jwtOAuth) RegisterSigners(ctx context.Context, req *RegisterSignersReq) error {
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
	}

	for _, signer := range req.Signers {
		if !IsSignerAddress(signer) {
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
	if err := userPermCheck(ctx, req.User); err != nil {
		return false, fmt.Errorf("need admin prem or user %s does not match: %w", req.User, err)
	}

	addr := req.Signer
	if !IsSignerAddress(addr) {
		return false, fmt.Errorf("invalid protocol type: %v", addr.Protocol())
	}

	has, err := o.store.SignerExistInUser(addr, req.User)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (o *jwtOAuth) ListSigner(ctx context.Context, req *ListSignerReq) (ListSignerResp, error) {
	if err := userPermCheck(ctx, req.User); err != nil {
		return nil, fmt.Errorf("need admin prem or user %s does not match: %w", req.User, err)
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
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return fmt.Errorf("need admin prem: %w", err)
	}

	for _, signer := range req.Signers {
		if !IsSignerAddress(signer) {
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
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		return false, fmt.Errorf("need admin prem: %w", err)
	}

	addr := req.Signer
	if !IsSignerAddress(addr) {
		return false, fmt.Errorf("invalid protocol type: %v", addr.Protocol())
	}

	return o.store.HasSigner(addr)
}

func (o jwtOAuth) DelSigner(ctx context.Context, req *DelSignerReq) (bool, error) {
	addr := req.Signer
	err := permCheck(ctx, core.PermAdmin)
	if err != nil {
		if err := ownerOfSignerCheck(ctx, o.store, req.Signer); err != nil {
			return false, fmt.Errorf("need admin prem or %s ownership check error: %w", req.Signer, err)
		}
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

func IsSignerAddress(addr address.Address) bool {
	protocol := addr.Protocol()
	return protocol == address.SECP256K1 || protocol == address.BLS || protocol == address.Delegated
}

func permCheck(ctx context.Context, needPerm core.Permission) error {
	callerPerms, ok := core.CtxGetPerm(ctx)
	if !ok {
		return ErrorPermissionNotFound
	}

	for _, callerPerm := range callerPerms {
		if callerPerm == needPerm {
			return nil
		}
	}

	return ErrorPermissionDeny
}

func userPermCheck(ctx context.Context, username string) error {
	err := permCheck(ctx, core.PermAdmin)
	if err == nil {
		return nil
	}

	ctxUsername, ok := core.CtxGetName(ctx)
	if !ok {
		return ErrorUsernameNotFound
	}
	if ctxUsername == username {
		return nil
	}

	return ErrorPermissionDeny
}

func tokenPermCheck(ctx context.Context, token string) error {
	err := permCheck(ctx, core.PermAdmin)
	if err == nil {
		return nil
	}

	username, err := JwtUserFromToken(token)
	if err != nil {
		return fmt.Errorf("get user of token %s failed: %w", token, err)
	}

	ctxUsername, ok := core.CtxGetName(ctx)
	if !ok {
		return ErrorUsernameNotFound
	}
	if ctxUsername == username {
		return nil
	}

	return ErrorPermissionDeny
}

type minerOwnershipChecker interface {
	MinerExistInUser(maddr address.Address, userName string) (bool, error)
}

func ownerOfMinerCheck(ctx context.Context, checker minerOwnershipChecker, miner address.Address) error {
	userName, ok := core.CtxGetName(ctx)
	if !ok {
		return ErrorUsernameNotFound
	}
	has, err := checker.MinerExistInUser(miner, userName)
	if err != nil {
		return err
	}

	if !has {
		return ErrorPermissionDeny
	}

	return nil
}

type signerOwnershipChecker interface {
	SignerExistInUser(signer address.Address, userName string) (bool, error)
}

func ownerOfSignerCheck(ctx context.Context, checker signerOwnershipChecker, signer address.Address) error {
	userName, ok := core.CtxGetName(ctx)
	if !ok {
		return ErrorUsernameNotFound
	}
	has, err := checker.SignerExistInUser(signer, userName)
	if err != nil {
		return err
	}

	if !has {
		return ErrorPermissionDeny
	}

	return nil
}
