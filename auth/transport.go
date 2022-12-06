package auth

import (
	"time"

	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/storage"
)

type VerifyRequest struct {
	Token string `form:"token" binding:"required"`
}
type VerifyResponse = JWTPayload

type GenTokenRequest struct {
	Name  string `form:"name" json:"name" binding:"required"`
	Perm  string `form:"perm" json:"perm"`
	Extra string `form:"extra" json:"extra"`
}

type GenTokenResponse struct {
	Token string `json:"token"`
}

type GetTokenRequest struct {
	Token string `form:"token" json:"token"`
	Name  string `form:"name" json:"name"`
}

type RemoveTokenRequest struct {
	Token string `form:"token" json:"token" binding:"required"`
}

type RecoverTokenRequest struct {
	Token string `form:"token" json:"token" binding:"required"`
}

type GetTokensRequest struct {
	*core.Page
}

func NewListUsersRequest(skip, limit int64, state int) *ListUsersRequest {
	return &ListUsersRequest{
		Page: &core.Page{
			Skip:  skip,
			Limit: limit,
		},
		State: state,
	}
}

type ListUsersRequest struct {
	*core.Page
	State int `form:"state" json:"state"`
}

type ListUsersResponse = []*OutputUser

type GetTokensResponse = []*TokenInfo

type GetUserRateLimitsReq struct {
	Id   string `form:"id"`
	Name string `form:"name" binding:"required"`
}

type DelUserRateLimitReq struct {
	Name string `form:"name"`
	Id   string `form:"id"`
}

type (
	GetUserRateLimitResponse []*storage.UserRateLimit
	UpsertUserRateLimitReq   storage.UserRateLimit
)

type CreateUserRequest struct {
	Name    string         `form:"name" binding:"required"`
	Comment *string        `form:"comment"`
	State   core.UserState `form:"state"` // 0: disable, 1: enable
}
type CreateUserResponse = OutputUser

type UpdateUserRequest struct {
	Name    string         `form:"name"`
	Comment *string        `form:"comment"`
	State   core.UserState `form:"state"`
}

type OutputUser struct {
	Id         string         `json:"id"`
	Name       string         `json:"name"`
	Comment    string         `json:"comment"`
	State      core.UserState `json:"state"`
	CreateTime int64          `json:"createTime"`
	UpdateTime int64          `json:"updateTime"`
	// the field `Miners` is used for compound api `ListUserWithMiners`
	// which calls 'listuser' and for each 'user' calls 'listminers'
	Miners []*OutputMiner `json:"-"`
}

type VerifyUsersReq struct {
	Names []string `form:"names" binding:"required"`
}

type GetUserRequest struct {
	Name string `form:"name" binding:"required"`
}

type HasUserRequest struct {
	Name string `form:"name" binding:"required"`
}

type DeleteUserRequest struct {
	Name string `form:"name" binding:"required"`
}

type RecoverUserRequest struct {
	Name string `form:"name" binding:"required"`
}

type GetUserByMinerRequest struct {
	// todo make miner tobe address
	Miner string `form:"miner" binding:"required"`
}

func (ls GetUserRateLimitResponse) MatchedLimit(service, api string) *storage.UserRateLimit {
	// just returns root matched limit currently
	// todo: returns most matched limit
	for _, l := range ls {
		if l.Service == "" && l.API == "" {
			return l
		}
	}
	return nil
}

type UpsertMinerReq struct {
	User       string `binding:"required"`
	Miner      string `binding:"required"`
	OpenMining *bool  `binding:"required"`
}

type HasMinerRequest struct {
	Miner string `form:"miner" binding:"required"`
}

type MinerExistInUserRequest struct {
	Miner string `form:"miner"`
	User  string `form:"user"`
}

type ListMinerReq struct {
	User string `form:"user" binding:"required"`
}

type OutputMiner struct {
	Miner, User          string
	OpenMining           bool
	CreatedAt, UpdatedAt time.Time
}
type ListMinerResp []*OutputMiner

type DelMinerReq struct {
	Miner string `json:"miner"`
}

// type definitions for signer
type RegisterSignersReq struct {
	User    string
	Signers []string
}

type UnregisterSignersReq struct {
	User    string
	Signers []string
}

type SignerExistInUserReq struct {
	Signer string `form:"signer"`
	User   string `form:"user"`
}

type ListSignerReq struct {
	User string `form:"user"`
}

type GetUserBySignerReq struct {
	Signer string `form:"signer"`
}

type HasSignerReq struct {
	Signer string `form:"signer"`
}

type DelSignerReq struct {
	Signer string `json:"signer"`
}

type OutputSigner struct {
	Signer, User         string
	CreatedAt, UpdatedAt time.Time
}
type ListSignerResp []*OutputSigner
