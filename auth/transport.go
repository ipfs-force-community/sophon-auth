package auth

import (
	"github.com/filecoin-project/venus-auth/core"
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

type CreateUserRequest struct {
	Name       string          `form:"name" binding:"required"`
	Miner      string          `form:"miner"` // miner address f01234
	Comment    string          `form:"comment"`
	State      int             `form:"state"` // 0: disable, 1: enable
	SourceType core.SourceType `form:"sourceType"`
}
type CreateUserResponse = OutputUser

type UpdateUserRequest struct {
	KeySum     core.KeyCode    `form:"keySum"` // keyCode Sum
	Name       string          `form:"name"`
	Miner      string          `form:"miner"`      // keyCode:1
	Comment    string          `form:"comment"`    // keyCode:2
	State      int             `form:"state"`      // keyCode:4
	SourceType core.SourceType `form:"sourceType"` // keyCode:8
}

type OutputUser struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	Miner      string          `json:"miner"` // miner address f01234
	SourceType core.SourceType `json:"sourceType"`
	Comment    string          `json:"comment"`
	State      int             `json:"state"`
	CreateTime int64           `json:"createTime"`
	UpdateTime int64           `json:"updateTime"`
}
