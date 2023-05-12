package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdaptOldStrategy(t *testing.T) {
	perms := AdaptOldStrategy(PermAdmin)
	assert.Equal(t, perms, []Permission{PermAdmin, PermSign, PermWrite, PermRead})
}

func TestWithPerm(t *testing.T) {
	ctx := CtxWithPerm(context.Background(), PermAdmin)
	callerPerms, ok := CtxGetPerm(ctx)
	assert.Equal(t, true, ok)
	assert.Equal(t, AdaptOldStrategy(PermAdmin), callerPerms)
}
