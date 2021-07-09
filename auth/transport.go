package auth

import (
	"github.com/filecoin-project/go-address"
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
func NewListAccountsRequest(skip, limit int64, sourceType core.SourceType, state int, keySum core.KeyCode) *ListAccountsRequest {
	return &ListAccountsRequest{
		Page: &core.Page{
			Skip:  skip,
			Limit: limit,
		},
		SourceType: sourceType,
		State:      state,
		KeySum:     keySum,
	}
}

type ListAccountsRequest struct {
	*core.Page
	SourceType core.SourceType `form:"sourceType" json:"sourceType"` // keyCode:1
	State      int             `form:"state" json:"state"`           // keyCode: 2
	KeySum     core.KeyCode    `form:"keySum"`                       // keyCode sum
}

type ListAccountsResponse = []*OutputAccount

type GetTokensResponse = []*TokenInfo

type CreateAccountRequest struct {
	Name       string           `form:"name" binding:"required"`
	Miner      string           `form:"miner"` // miner address f01234
	Comment    string           `form:"comment"`
	State      int              `form:"state"` // 0: disable, 1: enable
	SourceType core.SourceType  `form:"sourceType"`
	ReqLimit   storage.ReqLimit `form:"reqLimitAmount"`
}
type CreateAccountResponse = OutputAccount

type UpdateAccountRequest struct {
	KeySum core.KeyCode `form:"keySum"` // keyCode Sum
	Name   string       `form:"name"`
	// todo make miner tobe address
	Miner      string           `form:"miner"`      // keyCode:1
	Comment    string           `form:"comment"`    // keyCode:2
	State      int              `form:"state"`      // keyCode:4
	SourceType core.SourceType  `form:"sourceType"` // keyCode:8
	ReqLimit   storage.ReqLimit `form:"reqLimit"`   // keyCode:16
}

type OutputAccount struct {
	Id         string           `json:"id"`
	Name       string           `json:"name"`
	Miner      address.Address  `json:"miner"` // miner address f01234
	SourceType core.SourceType  `json:"sourceType"`
	ReqLimit   storage.ReqLimit `form:"reqLimit"`
	Comment    string           `json:"comment"`
	State      int              `json:"state"`
	CreateTime int64            `json:"createTime"`
	UpdateTime int64            `json:"updateTime"`
}

type GetAccountRequest struct {
	Name string `form:"name"`
}

type HasMinerRequest struct {
	// todo make miner tobe address
	Miner string `form:"miner"`
}

type GetMinerRequest struct {
	// todo make miner tobe address
	Miner string `form:"miner"`
}
