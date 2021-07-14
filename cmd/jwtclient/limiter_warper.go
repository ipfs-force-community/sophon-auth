package jwtclient

import (
	"errors"
	va "github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/ipfs-force-community/metrics/ratelimit"
)

type limitFinder struct {
	*JWTClient
}

var _ ratelimit.ILimitFinder = (*limitFinder)(nil)

var errNilJwtClient = errors.New("jwt client is nil")

func WarpLimitFinder(client *JWTClient) ratelimit.ILimitFinder {
	return &limitFinder{JWTClient: client}
}

func (l *limitFinder) GetUserLimit(name string) (*ratelimit.Limit, error) {
	if l.JWTClient == nil {
		return nil, errNilJwtClient
	}
	res, err := l.GetUser(&va.GetUserRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return &ratelimit.Limit{
		Account: res.Name, Cap: res.ReqLimit.Cap, Duration: res.ReqLimit.ResetDur}, nil
}

func (l *limitFinder) ListUserLimits() ([]*ratelimit.Limit, error) {
	if l.JWTClient == nil {
		return nil, errNilJwtClient
	}
	const PageSize = 20

	var limits = make([]*ratelimit.Limit, 0, PageSize*2)

	req := &va.ListUsersRequest{
		Page:       &core.Page{Skip: 0, Limit: PageSize},
		SourceType: 0, State: 0, KeySum: 0}

	for int64(len(limits)) == req.Skip {
		res, err := l.ListUsers(req)
		if err != nil {
			return nil, err
		}
		for _, u := range res {
			limits = append(limits,
				&ratelimit.Limit{Account: u.Name, Cap: u.ReqLimit.Cap, Duration: u.ReqLimit.ResetDur})
		}

		req.Skip += PageSize
	}
	return limits, nil

}
