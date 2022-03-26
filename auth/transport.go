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

type RemoveTokenRequest struct {
	Token string `form:"token" json:"token" binding:"required"`
}

type GetTokensRequest struct {
	*core.Page
}

// @sourceType: keyCode 1
// @state: keyCode 2
// @keySum: request params code sum, enum 1 2 4 8, to multi-select
func NewListUsersRequest(skip, limit int64, sourceType core.SourceType, state int, keySum core.KeyCode) *ListUsersRequest {
	return &ListUsersRequest{
		Page: &core.Page{
			Skip:  skip,
			Limit: limit,
		},
		SourceType: sourceType,
		State:      state,
		KeySum:     keySum,
	}
}

type ListUsersRequest struct {
	*core.Page
	SourceType core.SourceType `form:"sourceType" json:"sourceType"` // keyCode:1
	State      int             `form:"state" json:"state"`           // keyCode: 2
	KeySum     core.KeyCode    `form:"keySum"`                       // keyCode sum
}

type ListUsersResponse = []*OutputUser

type GetTokensResponse = []*TokenInfo

type GetUserRateLimitsReq struct {
	Id   string `form:"id"`
	Name string `form:"name"`
}

type DelUserRateLimitReq struct {
	Name string `form:"name"`
	Id   string `form:"id"`
}

type GetUserRateLimitResponse []*storage.UserRateLimit
type UpsertUserRateLimitReq storage.UserRateLimit

type CreateUserRequest struct {
	Name       string          `form:"name" binding:"required"`
	Comment    string          `form:"comment"`
	State      core.UserState  `form:"state"` // 0: disable, 1: enable
	SourceType core.SourceType `form:"sourceType"`
}
type CreateUserResponse = OutputUser

type UpdateUserRequest struct {
	KeySum core.KeyCode `form:"keySum"` // keyCode Sum
	Name   string       `form:"name"`
	// todo make miner tobe address
	Comment    string          `form:"comment"`    // keyCode:2
	State      core.UserState  `form:"state"`      // keyCode:4
	SourceType core.SourceType `form:"sourceType"` // keyCode:8
}

type OutputUser struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	SourceType core.SourceType `json:"sourceType"`
	Comment    string          `json:"comment"`
	State      core.UserState  `json:"state"`
	CreateTime int64           `json:"createTime"`
	UpdateTime int64           `json:"updateTime"`
}

type GetUserRequest struct {
	Name string `form:"name"`
}

type HasMinerRequest struct {
	// todo make miner tobe address
	Miner string `form:"miner"`
}

type GetUserByMinerRequest struct {
	// todo make miner tobe address
	Miner string `form:"miner"`
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
	User, Miner string
}

type ListMinerReq struct {
	User string `form:"user"`
}

type OutputMiner struct {
	Miner, User          string
	CreatedAt, UpdatedAt time.Time
}

type ListMinerResp []*OutputMiner

type DelMinerReq struct {
	Miner string `json:"miner"`
}
