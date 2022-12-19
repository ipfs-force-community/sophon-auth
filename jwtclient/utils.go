package jwtclient

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/venus-auth/core"
)

var (
	// ErrorPermissionDeny is the error message returned when a user does not have permission to perform an action
	ErrorPermissionDeny = fmt.Errorf("permission deny")

	// ErrorUserNotFound is the error message returned when a user is not found in context
	ErrorUserNotFound = fmt.Errorf("user not found")
)

// checkPermissionByUser check weather the user has admin permission or is match the username passed in
func CheckPermissionByName(ctx context.Context, name string) error {
	if auth.HasPerm(ctx, []auth.Permission{}, core.PermAdmin) {
		return nil
	}
	user, exist := CtxGetName(ctx)
	if !exist || user != name {
		return ErrorPermissionDeny
	}
	return nil
}

func CheckPermissionBySigner(ctx context.Context, client IAuthClient, addrs ...address.Address) error {
	if auth.HasPerm(ctx, []auth.Permission{}, core.PermAdmin) {
		return nil
	}
	user, exist := CtxGetName(ctx)
	if !exist {
		return ErrorUserNotFound
	}

	for _, wAddr := range addrs {
		ok, err := client.SignerExistInUser(ctx, user, wAddr)
		if err != nil {
			return fmt.Errorf("check signer exist in user fail %s failed when check permission: %s", wAddr.String(), err)
		}
		if !ok {
			return ErrorPermissionDeny
		}
	}
	return nil
}

func CheckPermissionByMiner(ctx context.Context, client IAuthClient, addrs ...address.Address) error {
	if auth.HasPerm(ctx, []auth.Permission{}, core.PermAdmin) {
		return nil
	}
	user, exist := CtxGetName(ctx)
	if !exist {
		return ErrorUserNotFound
	}
	for _, mAddr := range addrs {
		ok, err := client.MinerExistInUser(ctx, user, mAddr)
		if err != nil {
			return fmt.Errorf("check miner exist in user fail %s failed when check permission: %s", mAddr.String(), err)
		}
		if !ok {
			return ErrorPermissionDeny
		}
	}
	return nil
}
