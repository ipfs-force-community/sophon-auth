package auth

type VerifyRequest struct {
	Token string `form:"token" binding:"required"`
}
type VerifyResponse = JWTPayload

type GenTokenRequest struct {
	Name  string `form:"name" binding:"required"`
	Perm  string `form:"perm"`
	Extra string `form:"extra"`
}

type GenTokenResponse struct {
	Token string `json:"token"`
}

type RemoveTokenRequest struct {
	Token string `form:"token" binding:"required"`
}

type GetTokensRequest struct {
	PageIndex int64 `form:"pageIndex"`
	PageSize  int64 `form:"pageSize"`
}

func (o *GetTokensRequest) GetPageIndex() int64 {
	if o.PageIndex < 1 {
		return 1
	}
	return o.PageIndex
}
func (o *GetTokensRequest) GetPageSize() int64 {
	if o.PageSize < 1 {
		return 1
	} else if o.PageSize > 100 {
		return 100
	}
	return o.PageSize
}

type GetTokensResponse = []*TokenInfo
