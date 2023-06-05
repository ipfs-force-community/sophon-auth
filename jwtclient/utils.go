package jwtclient

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"

	"github.com/ipfs-force-community/sophon-auth/core"
)

var (
	// ErrorPermissionDeny is the error message returned when a user does not have permission to perform an action
	ErrorPermissionDeny = fmt.Errorf("permission deny")
)

// CheckPermissionByName check weather the user has admin permission or is match the username passed in
func CheckPermissionByName(ctx context.Context, username string) error {
	if core.HasPerm(ctx, []core.Permission{}, core.PermAdmin) {
		return nil
	}

	user, exist := core.CtxGetName(ctx)
	if !exist {
		return fmt.Errorf("there is no accountKey in the request")
	}

	if user != username {
		return fmt.Errorf("user %s in the request doesn't match %s in the system: %w", username, user, ErrorPermissionDeny)
	}

	return nil
}

func CheckPermissionBySigner(ctx context.Context, client IAuthClient, signers ...address.Address) error {
	if core.HasPerm(ctx, []core.Permission{}, core.PermAdmin) {
		return nil
	}

	user, exist := core.CtxGetName(ctx)
	if !exist {
		return fmt.Errorf("there is no accountKey in the request")
	}

	for _, addr := range signers {
		ok, err := client.SignerExistInUser(ctx, user, addr)
		if err != nil {
			return fmt.Errorf("check signer %s exist in user %s failed when check permission: %w", addr.String(), user, err)
		}
		if !ok {
			return fmt.Errorf("signer %s not exist in user %s: %w", addr.String(), user, ErrorPermissionDeny)
		}
	}
	return nil
}

func CheckPermissionByMiner(ctx context.Context, client IAuthClient, miners ...address.Address) error {
	if core.HasPerm(ctx, []core.Permission{}, core.PermAdmin) {
		return nil
	}

	user, exist := core.CtxGetName(ctx)
	if !exist {
		return fmt.Errorf("there is no accountKey in the request")
	}

	for _, mAddr := range miners {
		ok, err := client.MinerExistInUser(ctx, user, mAddr)
		if err != nil {
			return fmt.Errorf("check miner %s exist in user %s failed when check permission: %w", mAddr.String(), user, err)
		}
		if !ok {
			return fmt.Errorf("miner %s not exist in user %s: %w", mAddr.String(), user, ErrorPermissionDeny)
		}
	}
	return nil
}
