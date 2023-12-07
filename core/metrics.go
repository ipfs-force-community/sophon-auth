package core

import (
	"github.com/ipfs-force-community/metrics"
	"go.opencensus.io/tag"
)

var (
	emptyUnit = ""

	VerifyStateFailed  = "failed"
	VerifyStateSuccess = "success"

	TagPerm        = tag.MustNewKey("perm")
	TagUserState   = tag.MustNewKey("user_state")
	TagTokenName   = tag.MustNewKey("token_name")
	TagUserName    = tag.MustNewKey("user_name")
	TagVerifyState = tag.MustNewKey("verify_state")
)

var (
	TokenGauge         = metrics.NewInt64WithCategory("token/amount", "amount of token", emptyUnit)
	UserGauge          = metrics.NewInt64WithCategory("user/amount", "amount of user", emptyUnit)
	TokenVerifyCounter = metrics.NewCounter("token/verify", "amount of token verify", TagPerm, TagVerifyState)
	ApiState           = metrics.NewInt64("api/state", "api service state. 0: down, 1: up", emptyUnit)
)
