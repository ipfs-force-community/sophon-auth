package auth

import "github.com/filecoin-project/venus-auth/storage"

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
	Skip  int64 `form:"skip" json:"skip"`
	Limit int64 `form:"limit" json:"limit"`
}

func (o *GetTokensRequest) GetSkip() int64 {
	if o.Skip < 0 {
		o.Skip = 0
	}
	return o.Skip
}
func (o *GetTokensRequest) GetLimit() int64 {
	if o.Limit < 0 || o.Limit > 20 {
		o.Limit = 20
	}
	return o.Limit
}

type ListUsersRequest struct {
	Skip  int64 `form:"skip" json:"skip"`
	Limit int64 `form:"limit" json:"limit"`
}

func (o *ListUsersRequest) GetSkip() int64 {
	if o.Skip < 0 {
		o.Skip = 0
	}
	return o.Skip
}
func (o *ListUsersRequest) GetLimit() int64 {
	if o.Limit < 0 || o.Limit > 20 {
		o.Limit = 20
	}
	return o.Limit
}

type GetTokensResponse = []*TokenInfo

type ListUsersResponse = []*storage.User
