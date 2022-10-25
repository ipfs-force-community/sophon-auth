package jwtclient

import (
	"errors"

	"github.com/ipfs-force-community/metrics/ratelimit"
)

type limitFinder struct {
	*AuthClient
}

var _ ratelimit.ILimitFinder = (*limitFinder)(nil)

var errNilJwtClient = errors.New("jwt client is nil")

func WarpLimitFinder(client *AuthClient) ratelimit.ILimitFinder {
	return &limitFinder{AuthClient: client}
}

func (l *limitFinder) GetUserLimit(name, service, api string) (*ratelimit.Limit, error) {
	if l.AuthClient == nil {
		return nil, errNilJwtClient
	}

	res, err := l.GetUserRateLimit(name, "")
	if err != nil {
		return nil, err
	}

	limit := &ratelimit.Limit{Account: name, Cap: 0, Duration: 0}
	if l := res.MatchedLimit(service, api); l != nil {
		limit.Cap = l.ReqLimit.Cap
		limit.Duration = l.ReqLimit.ResetDur
	}

	return limit, nil
}
