package core

import (
	"context"
	"gotest.tools/assert"
	"testing"
)

func TestAdaptOldStrategy(t *testing.T) {
	perms := AdaptOldStrategy(PermAdmin)
	assert.DeepEqual(t, perms, []Permission{PermAdmin, PermSign, PermWrite, PermRead})
}

func TestWithPerm(t *testing.T) {
	ctx := WithPerm(context.Background(), PermAdmin)
	callerPerms, ok := ctx.Value(PermCtxKey).([]Permission)
	if !ok {
		t.Fatal()
	}
	t.Log(callerPerms)
}
