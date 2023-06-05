package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdaptOldStrategy(t *testing.T) {
	perms := AdaptOldStrategy(PermAdmin)
	assert.Equal(t, perms, []Permission{PermRead, PermWrite, PermSign, PermAdmin})
}

func TestWithPerm(t *testing.T) {
	for _, perm := range PermArr {
		ctx := CtxWithPerm(context.Background(), perm)
		callerPerms, ok := CtxGetPerm(ctx)
		assert.Equal(t, true, ok)
		assert.Equal(t, AdaptOldStrategy(perm), callerPerms)

		ctx = CtxWithPerms(context.Background(), AdaptOldStrategy(perm))
		callerPerms, ok = CtxGetPerm(ctx)
		assert.Equal(t, true, ok)
		assert.Equal(t, AdaptOldStrategy(perm), callerPerms)
	}
}
