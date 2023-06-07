// stm: #unit
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-address"

	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/storage"
)

var (
	adminCtx = core.CtxWithPerm(context.Background(), core.PermAdmin)
	signCtx  = core.CtxWithPerm(context.Background(), core.PermSign)
	readCtx  = core.CtxWithPerm(context.Background(), core.PermRead)
)

func signPermAndRandomUsername() context.Context {
	ctx := core.CtxWithPerm(context.Background(), core.PermSign)
	return core.CtxWithName(ctx, time.Now().String())
}

func TestJwt(t *testing.T) {
	limitStrs := `[{"Id":"794fc9a4-2b80-4503-835a-7e8e27360b3d","Name":"test_user_01","Service":"","API":"","ReqLimit":{"Cap":10,"ResetDur":120000000000}},{"Id":"252f581e-cbd2-4a61-a517-0b7df65013aa","Name":"test_user_02","Service":"","API":"","ReqLimit":{"Cap":10,"ResetDur":72000000000000}}]`
	var originLimits []*storage.UserRateLimit

	if err := json.Unmarshal([]byte(limitStrs), &originLimits); err != nil {
		t.Fatalf("initialize origin Ratelimit failed:%s", err.Error())
	}

	userMiners := map[string][]string{
		"test_user_001": {"t01000", "t01002", "t01003"},
		"test_user_002": {"t01004", "t01005", "t01006"},
		"test_user_003": {"t01007", "t01008", "t01009"},
	}

	// stm: @VENUSAUTH_JWT_GENERATE_TOKEN_001
	t.Run("generate token", testGenerateToken)
	// stm: @VENUSAUTH_JWT_VERIFY_TOKEN_001, @VENUSAUTH_JWT_VERIFY_TOKEN_002
	t.Run("verify token", testVerifyToken)
	// stm: @VENUSAUTH_JWT_GET_TOKEN_001, @VENUSAUTH_JWT_GET_TOKEN_002
	t.Run("get token", testGetToken)
	// stm: @VENUSAUTH_JWT_GET_TOKEN_BY_NAME_001, @VENUSAUTH_JWT_GET_TOKEN_BY_NAME_002
	t.Run("get token by name", testGetTokenByName)
	// stm: @VENUSAUTH_JWT_TOKENS_001
	t.Run("list all tokens", testTokenList)
	// stm: @VENUSAUTH_JWT_REMOVE_TOKEN_001, @VENUSAUTH_JWT_RECOVER_TOKEN_001, @VENUSAUTH_JWT_RECOVER_TOKEN_003
	t.Run("remove and recover tokens", testRemoveAndRecoverToken)
	// Features about users
	// stm: @VENUSAUTH_JWT_CREATE_USER_001, @VENUSAUTH_JWT_CREATE_USER_003
	t.Run("test create user", func(t *testing.T) { testCreateUser(t, userMiners) })
	// stm: @VENUSAUTH_JWT_GET_USER_001, @VENUSAUTH_JWT_HAS_USER_001, @VENUSAUTH_JWT_GET_USER_002
	t.Run("test get user", func(t *testing.T) { testGetUser(t, userMiners) })
	// stm: @VENUSAUTH_JWT_VERIFY_USERS_001
	t.Run("test verify user", func(t *testing.T) { testVerifyUsers(t, userMiners) })
	// stm: @VENUSAUTH_JWT_LIST_USERS_001
	t.Run("test list user", func(t *testing.T) { testListUser(t, userMiners) })
	// stm: @VENUSAUTH_JWT_UPDATE_USER_001, @VENUSAUTH_JWT_UPDATE_USER_002
	t.Run("test update user", func(t *testing.T) { testUpdateUser(t, userMiners) })
	// stm: @VENUSAUTH_JWT_DELETE_USER_001, @VENUSAUTH_JWT_RECOVER_USER_001, @VENUSAUTH_JWT_RECOVER_USER_002, @VENUSAUTH_JWT_RECOVER_USER_003
	t.Run("test delete and recover user", func(t *testing.T) { testDeleteAndRecoverUser(t, userMiners) })
	// Features about miners
	// stm: @VENUSAUTH_JWT_UPSERT_MINER_001, @VENUSAUTH_JWT_UPSERT_MINER_002
	t.Run("test upsert miner", func(t *testing.T) { testUpsertMiner(t, userMiners) })
	// stm: @VENUSAUTH_JWT_LIST_MINERS_001
	t.Run("test list miner", func(t *testing.T) { testListMiner(t, userMiners) })
	// stm: @VENUSAUTH_JWT_HAS_MINER_001, @VENUSAUTH_JWT_HAS_MINER_002
	t.Run("test miner exist user", func(t *testing.T) { testMinerExistInMiner(t, userMiners) })
	// stm: @VENUSAUTH_JWT_GET_USER_BY_MINER_001, @VENUSAUTH_JWT_GET_USER_BY_MINER_002, @VENUSAUTH_JWT_GET_USER_BY_MINER_003
	t.Run("test get user by miner", func(t *testing.T) { testGetUserByMiner(t, userMiners) })
	// stm: @VENUSAUTH_JWT_DELETE_MINER_001, @VENUSAUTH_JWT_DELETE_MINER_002
	t.Run("test delete miner", func(t *testing.T) { testDeleteMiner(t, userMiners) })

	// Features about signers
	userSigners := map[string][]string{
		"test_user_001": {"t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha", "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"},
		"test_user_002": {"t3r47fkdzfmtex5ic3jnwlzc7bkpbj7s4d6limyt4f57t3cuqq5nuvhvwv2cu2a6iga2s64vjqcxjqiezyjooq", "t1uqtvvwkkfkkez52ocnqe6vg74qewiwja4t2tiba"},
		"test_user_003": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
		"test_user_004": {"f410ffhviosf3a4ymvgm3sxrugj2qkir33qnlhqjrauq"},
	}
	t.Run("test register signer", func(t *testing.T) { testRegisterSigner(t, userSigners) })
	t.Run("test signer exist in user", func(t *testing.T) { testSignerExistInUser(t, userSigners) })
	t.Run("test list signer", func(t *testing.T) { testListSigner(t, userSigners) })
	t.Run("test has signer", func(t *testing.T) { testHasSigner(t, userSigners) })
	t.Run("test get user by signer", func(t *testing.T) { testGetUserBySigner(t, userSigners) })
	t.Run("test unregister signer", func(t *testing.T) { testUnregisterSigner(t, userSigners) })
	t.Run("test delete signer", func(t *testing.T) { testDeleteSigner(t, userSigners) })

	// Features about rate limits
	// stm: @VENUSAUTH_JWT_UPSERT_USER_RATE_LIMITS_001
	t.Run("test upsert rate limit", func(t *testing.T) { testUpsertUserRateLimit(t, userMiners, originLimits) })
	t.Run("test get rate limit", func(t *testing.T) { testGetUserRateLimits(t, userMiners, originLimits) })
	// stm: @VENUSAUTH_JWT_DELETE_USER_RATE_LIMITS_001
	t.Run("test delete rate limit", func(t *testing.T) { testDeleteUserRateLimits(t, userMiners, originLimits) })
}

func testGenerateToken(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "admin",
		Extra: "",
	}

	_, err := jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.NotNil(t, err)
	require.Contains(t, err.Error(), "token must be based on an existing user")

	createUserReq := &CreateUserRequest{
		Name:  "test-token-01",
		State: 0,
	}
	resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-01", resp.Name)
	token, err := jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(strings.Split(token, ".")))

	_, err = jwtOAuthInstance.GenerateToken(signCtx, pl1)
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testVerifyToken(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// First generate a token
	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "admin",
		Extra: "",
	}

	createUserReq := &CreateUserRequest{
		Name:  "test-token-01",
		State: 0,
	}
	resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-01", resp.Name)
	token, err := jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.Nil(t, err)

	// Verify a valid token
	payload1, err := jwtOAuthInstance.Verify(readCtx, token)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(payload1, pl1))

	// Try to verify an invalid token
	invalidToken := "I'm just an invalid token"
	_, err = jwtOAuthInstance.Verify(readCtx, invalidToken)
	assert.NotNil(t, err)

	// with ctx no perm
	_, err = jwtOAuthInstance.Verify(context.Background(), token)
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))
}

func testGetToken(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// First generate a token
	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "admin",
		Extra: "",
	}

	createUserReq := &CreateUserRequest{
		Name:  "test-token-01",
		State: 0,
	}
	resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-01", resp.Name)
	token, err := jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.Nil(t, err)

	// Get token
	tokenInfo1, err := jwtOAuthInstance.GetToken(adminCtx, token)
	assert.Nil(t, err)
	assert.Equal(t, pl1.Name, tokenInfo1.Name)
	assert.Equal(t, pl1.Perm, tokenInfo1.Perm)
	// Try to get invalid token
	invalidToken := "I'm just an invalid token"
	_, err = jwtOAuthInstance.GetToken(adminCtx, invalidToken)
	assert.NotNil(t, err)

	// with ctx no perm
	_, err = jwtOAuthInstance.GetToken(context.Background(), token)
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	_, err = jwtOAuthInstance.GetToken(signCtx, invalidToken)
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testGetTokenByName(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// First generate a token
	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "admin",
		Extra: "",
	}

	createUserReq := &CreateUserRequest{
		Name:  "test-token-01",
		State: 0,
	}
	resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-01", resp.Name)
	token, err := jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.Nil(t, err)
	userCtx := ctxWithUserNameAndAdminPerm(pl1.Name)

	validPermTest := func(ctx context.Context) {
		// Get token by name
		tokenInfoList1, err := jwtOAuthInstance.GetTokenByName(ctx, "test-token-01")
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tokenInfoList1))
		assert.Equal(t, token, tokenInfoList1[0].Token)

	}
	invalidPermTest := func(ctx context.Context, expect error) {
		_, err := jwtOAuthInstance.GetTokenByName(ctx, "test-token-01")
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	validPermTest(adminCtx)
	validPermTest(userCtx)
	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)
	invalidPermTest(context.Background(), ErrorUsernameNotFound)

	// Try to get token by wrong name
	tokenInfoInvalid, err := jwtOAuthInstance.GetTokenByName(adminCtx, "invalid_name")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tokenInfoInvalid))
}

func testTokenList(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// First generate two tokens
	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "admin",
		Extra: "",
	}
	pl2 := &JWTPayload{
		Name:  "test-token-02",
		Perm:  "admin",
		Extra: "",
	}

	createUserReq := &CreateUserRequest{
		Name:  "test-token-01",
		State: 0,
	}
	resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-01", resp.Name)
	_, err = jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.Nil(t, err)
	createUserReq = &CreateUserRequest{
		Name:  "test-token-02",
		State: 0,
	}
	resp, err = jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-02", resp.Name)
	_, err = jwtOAuthInstance.GenerateToken(adminCtx, pl2)
	assert.Nil(t, err)

	allTokenInfos, err := jwtOAuthInstance.Tokens(adminCtx, 0, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allTokenInfos))
	// with skip or limit
	allTokenInfos, err = jwtOAuthInstance.Tokens(adminCtx, 1, 10)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))

	allTokenInfos, err = jwtOAuthInstance.Tokens(adminCtx, 0, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))

	allTokenInfos, err = jwtOAuthInstance.Tokens(adminCtx, 2, 10)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(allTokenInfos))

	// with ctx no perm
	_, err = jwtOAuthInstance.Tokens(context.Background(), 0, 2)
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	_, err = jwtOAuthInstance.Tokens(signCtx, 0, 2)
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testRemoveAndRecoverToken(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// First generate a token
	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "read",
		Extra: "",
	}

	createUserReq := &CreateUserRequest{
		Name:  "test-token-01",
		State: 0,
	}
	resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	assert.Nil(t, err)
	assert.Equal(t, "test-token-01", resp.Name)
	token, err := jwtOAuthInstance.GenerateToken(adminCtx, pl1)
	assert.Nil(t, err)

	validPermTest := func(ctx context.Context) {

		// token is usable.
		err = jwtOAuthInstance.RecoverToken(ctx, token)
		assert.Error(t, err)

		// Remove a token
		err = jwtOAuthInstance.RemoveToken(ctx, token)
		assert.Nil(t, err)

		_, err = jwtOAuthInstance.Verify(ctx, token)
		assert.NotNil(t, err)

		tokenInfoList1, err := jwtOAuthInstance.GetTokenByName(ctx, "test-token-01")
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tokenInfoList1))

		// Recover a token
		err = jwtOAuthInstance.RecoverToken(ctx, token)
		assert.Nil(t, err)
		payload1, err := jwtOAuthInstance.Verify(ctx, token)
		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(payload1, pl1))
		allTokenInfos, err := jwtOAuthInstance.Tokens(adminCtx, 0, 2)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(allTokenInfos))
	}

	invalidPermTest := func(ctx context.Context, expect error) {
		// Remove a token
		err := jwtOAuthInstance.RemoveToken(ctx, token)
		assert.Equal(t, expect, errors.Unwrap(err))

		// Recover a token
		err = jwtOAuthInstance.RecoverToken(ctx, token)
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	userCtx := ctxWithUserNameAndAdminPerm(pl1.Name)

	validPermTest(userCtx)
	validPermTest(adminCtx)

	invalidPermTest(context.Background(), ErrorUsernameNotFound)
	invalidPermTest(signCtx, ErrorUsernameNotFound)
}

func createUsers(t *testing.T, userMiners map[string][]string) {
	ctx := adminCtx
	// Create 3 users
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:  userName,
			State: 0,
		}
		resp, err := jwtOAuthInstance.CreateUser(ctx, createUserReq)
		assert.Nil(t, err)
		assert.Equal(t, userName, resp.Name)
		assert.Equal(t, "", resp.Comment)
	}
}

func testCreateUser(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// with adminCtx admin perm
	existUserName := "test_user_001"
	comment := "test comment"
	// Create 3 users
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:    userName,
			Comment: &comment,
			State:   0,
		}
		resp, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
		assert.Nil(t, err)
		assert.Equal(t, userName, resp.Name)
		assert.Equal(t, "test comment", resp.Comment)
	}
	// Create duplicate user
	_, err := jwtOAuthInstance.CreateUser(adminCtx, &CreateUserRequest{Name: existUserName})
	assert.NotNil(t, err)

	// with ctx no perm
	_, err = jwtOAuthInstance.CreateUser(context.Background(), &CreateUserRequest{Name: "test_user_002"})
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	_, err = jwtOAuthInstance.CreateUser(signCtx, &CreateUserRequest{Name: "test_user_002"})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testGetUser(t *testing.T, userMiners map[string][]string) {
	existUserName := "test_user_001"
	invalidUserName := "invalid_name"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	validPermTest := func(ctx context.Context) {
		u, err := jwtOAuthInstance.GetUser(ctx, &GetUserRequest{Name: existUserName})
		assert.Nil(t, err)
		assert.Equal(t, u.Name, existUserName)
	}
	invalidPermTest := func(ctx context.Context, expect error) {
		_, err := jwtOAuthInstance.GetUser(ctx, &GetUserRequest{Name: existUserName})
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	userCtx := ctxWithUserNameAndAdminPerm(existUserName)
	validPermTest(adminCtx)
	validPermTest(userCtx)
	invalidPermTest(context.Background(), ErrorUsernameNotFound)
	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)

	exist, err := jwtOAuthInstance.HasUser(adminCtx, &HasUserRequest{Name: invalidUserName})
	assert.Nil(t, err)
	assert.False(t, exist)

}

func testVerifyUsers(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	usernames := make([]string, 0)
	for key := range userMiners {
		usernames = append(usernames, key)
	}

	// with ctx admin perm
	err := jwtOAuthInstance.VerifyUsers(adminCtx, &VerifyUsersReq{Names: usernames})
	assert.Nil(t, err)

	// with ctx no perm
	err = jwtOAuthInstance.VerifyUsers(context.Background(), &VerifyUsersReq{Names: usernames})
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	err = jwtOAuthInstance.VerifyUsers(signCtx, &VerifyUsersReq{Names: usernames})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testListUser(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	// with ctx admin perm
	allUserInfos, err := jwtOAuthInstance.ListUsers(adminCtx, &ListUsersRequest{
		Page:  &core.Page{},
		State: int(core.UserStateUndefined),
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allUserInfos))

	// with ctx no perm
	_, err = jwtOAuthInstance.ListUsers(context.Background(), &ListUsersRequest{
		Page:  &core.Page{},
		State: int(core.UserStateUndefined),
	})
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	_, err = jwtOAuthInstance.ListUsers(signCtx, &ListUsersRequest{
		Page:  &core.Page{},
		State: int(core.UserStateUndefined),
	})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testUpdateUser(t *testing.T, userMiners map[string][]string) {
	existUserName := "test_user_001"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	// Update a user
	comment := "New Comment"
	updateUserReq := &UpdateUserRequest{
		Name:    existUserName,
		Comment: &comment,
	}

	// with adminCtx admin perm
	err := jwtOAuthInstance.UpdateUser(adminCtx, updateUserReq)
	assert.Nil(t, err)
	// Then get this user
	outPutUser1, err := jwtOAuthInstance.GetUser(adminCtx, &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, "New Comment", outPutUser1.Comment)

	// invalid user name
	err = jwtOAuthInstance.UpdateUser(adminCtx, &UpdateUserRequest{})
	assert.Error(t, err)

	// with ctx no perm
	err = jwtOAuthInstance.UpdateUser(context.Background(), updateUserReq)
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	err = jwtOAuthInstance.UpdateUser(signCtx, updateUserReq)
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testDeleteAndRecoverUser(t *testing.T, userMiners map[string][]string) {
	existUserName := "test_user_001"
	invalidUserName := "invalid_name"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	// with ctx admin perm

	// Delete User
	err := jwtOAuthInstance.DeleteUser(adminCtx, &DeleteUserRequest{Name: existUserName})
	assert.Nil(t, err)
	// Then try to get this user
	_, err = jwtOAuthInstance.GetUser(adminCtx, &GetUserRequest{Name: existUserName})
	assert.NotNil(t, err)
	// And also list users now
	allUserInfos, err := jwtOAuthInstance.ListUsers(adminCtx, &ListUsersRequest{
		Page:  &core.Page{},
		State: int(core.UserStateUndefined),
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allUserInfos))

	// Try to delete non-existing users
	err = jwtOAuthInstance.DeleteUser(adminCtx, &DeleteUserRequest{Name: invalidUserName})
	assert.NotNil(t, err)

	// Recover user
	err = jwtOAuthInstance.RecoverUser(adminCtx, &RecoverUserRequest{Name: existUserName})
	assert.Nil(t, err)
	// Then get this user
	outPutUser1, err := jwtOAuthInstance.GetUser(adminCtx, &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, existUserName, outPutUser1.Name)

	// Try to recover an invalid user
	err = jwtOAuthInstance.RecoverUser(adminCtx, &RecoverUserRequest{Name: invalidUserName})
	assert.NotNil(t, err)

	// Try to recover a valid, but not deleted user
	err = jwtOAuthInstance.RecoverUser(adminCtx, &RecoverUserRequest{Name: existUserName})
	assert.NotNil(t, err)

	// with ctx no perm
	err = jwtOAuthInstance.DeleteUser(context.Background(), &DeleteUserRequest{Name: existUserName})
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))
	err = jwtOAuthInstance.DeleteUser(signCtx, &DeleteUserRequest{Name: existUserName})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))

	err = jwtOAuthInstance.RecoverUser(context.Background(), &RecoverUserRequest{Name: existUserName})
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))
	err = jwtOAuthInstance.RecoverUser(signCtx, &RecoverUserRequest{Name: existUserName})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func addUsersAndMiners(t *testing.T, userMiners map[string][]string) {
	ctx := adminCtx
	for userName, miners := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:  userName,
			State: 0,
		}
		// Create users.
		_, _ = jwtOAuthInstance.CreateUser(ctx, createUserReq)
		// Add miners
		openMining := true
		for _, minerID := range miners {
			mAddr, err := address.NewFromString(minerID)
			assert.Nil(t, err)
			ifCreate, err := jwtOAuthInstance.UpsertMiner(ctx, &UpsertMinerReq{
				User:       userName,
				Miner:      mAddr,
				OpenMining: &openMining,
			})
			assert.Nil(t, err)
			assert.True(t, ifCreate)
		}
	}
}

func testUpsertMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	// with adminCtx
	// error signer address
	_, _ = jwtOAuthInstance.CreateUser(adminCtx, &CreateUserRequest{
		Name:  "user_01",
		State: 1,
	})
	mAddr, _ := address.NewFromString("f01034")
	isCreate, err := jwtOAuthInstance.UpsertMiner(adminCtx, &UpsertMinerReq{User: "user_01", Miner: mAddr})
	assert.Nil(t, err)
	assert.True(t, isCreate)

	mAddr, err = address.NewFromString("f1mpvdqt2acgihevibd4greavlsfn3dfph5sckc2a")
	assert.Nil(t, err)
	_, err = jwtOAuthInstance.UpsertMiner(adminCtx, &UpsertMinerReq{
		User:  "user_01",
		Miner: mAddr,
	})
	assert.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid protocol type")

	// with signCtx
	_, err = jwtOAuthInstance.UpsertMiner(signCtx, &UpsertMinerReq{User: "user_01", Miner: mAddr})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testListMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	validUser1 := "test_user_001"
	user1Miners := []string{"t01000", "t01002", "t01003"}

	validPermTest := func(ctx context.Context) {
		// List miners
		resp, err := jwtOAuthInstance.ListMiners(ctx, &ListMinerReq{User: validUser1})
		assert.Nil(t, err)
		assert.Equal(t, len(user1Miners), len(resp))
		sort.Slice(resp, func(i, j int) bool { return resp[i].Miner.String() < resp[j].Miner.String() })
		for i := 0; i < len(user1Miners); i++ {
			assert.Equal(t, user1Miners[i], resp[i].Miner.String())
			assert.Equal(t, validUser1, resp[i].User)
			assert.Equal(t, true, resp[i].OpenMining)
		}
	}
	invalidPermTest := func(ctx context.Context, expect error) {
		_, err := jwtOAuthInstance.ListMiners(ctx, &ListMinerReq{User: validUser1})
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	userCtx := ctxWithUserNameAndAdminPerm(validUser1)
	validPermTest(userCtx)
	validPermTest(adminCtx)

	invalidPermTest(context.Background(), ErrorUsernameNotFound)
	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)
}

func testMinerExistInMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	validPermTest := func(ctx context.Context) {
		mAddr1, err := address.NewFromString("t01000")
		assert.Nil(t, err)
		// Miner Exist In Account
		exist, err := jwtOAuthInstance.MinerExistInUser(ctx, &MinerExistInUserRequest{Miner: mAddr1, User: "test_user_001"})
		assert.Nil(t, err)
		assert.True(t, exist)

		mAddr4, err := address.NewFromString("t01004")
		assert.Nil(t, err)
		exist, err = jwtOAuthInstance.MinerExistInUser(ctx, &MinerExistInUserRequest{Miner: mAddr4, User: "test_user_001"})
		assert.Nil(t, err)
		assert.False(t, exist)
	}
	invalidPermTest := func(ctx context.Context, expect error) {
		mAddr, _ := address.NewFromString("t01000")
		// Miner Exist In Account
		_, err := jwtOAuthInstance.MinerExistInUser(ctx, &MinerExistInUserRequest{Miner: mAddr, User: "test_user_001"})
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	userCtx := ctxWithUserNameAndAdminPerm("test_user_001")
	validPermTest(userCtx)
	validPermTest(adminCtx)

	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)
	invalidPermTest(context.Background(), ErrorUsernameNotFound)

	// Has Miner
	mAddr, _ := address.NewFromString("t01000")
	has, err := jwtOAuthInstance.HasMiner(adminCtx, &HasMinerRequest{Miner: mAddr})
	assert.Nil(t, err)
	assert.True(t, has)
	invalidMiner, _ := address.NewFromString("t01014")
	has, _ = jwtOAuthInstance.HasMiner(adminCtx, &HasMinerRequest{Miner: invalidMiner})
	assert.False(t, has)
	_, err = jwtOAuthInstance.HasMiner(signCtx, &HasMinerRequest{Miner: mAddr})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testGetUserByMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	// with ctx admin perm
	// Get User By Miner
	validUser1 := "test_user_001"
	miner := "t01002"
	mAddr, err := address.NewFromString(miner)
	assert.Nil(t, err)
	userInfo, err := jwtOAuthInstance.GetUserByMiner(adminCtx, &GetUserByMinerRequest{
		Miner: mAddr,
	})
	assert.Nil(t, err)
	assert.Equal(t, validUser1, userInfo.Name)

	// miner address not exist
	mAddr, err = address.NewFromString("f01989787")
	assert.Nil(t, err)
	_, err = jwtOAuthInstance.GetUserByMiner(adminCtx, &GetUserByMinerRequest{
		Miner: mAddr,
	})
	assert.Error(t, err)

	// with ctx no perm
	// Get User By Mine
	_, err = jwtOAuthInstance.GetUserByMiner(signCtx, &GetUserByMinerRequest{
		Miner: mAddr,
	})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testDeleteMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	user1 := "test_user_001"
	user1Miners := []string{"t01000", "t01002", "t01003"}
	invalidMiner := "t02000"

	validPermTest := func(ctx context.Context) {
		// Delete miner
		mAddr, err := address.NewFromString(user1Miners[0])
		assert.Nil(t, err)
		deleted, err := jwtOAuthInstance.DelMiner(ctx, &DelMinerReq{Miner: mAddr})
		assert.Nil(t, err)
		assert.True(t, deleted)
		// Then get this miner
		has, err := jwtOAuthInstance.HasMiner(adminCtx, &HasMinerRequest{Miner: mAddr})
		assert.Nil(t, err)
		assert.False(t, has)

		ok, err := jwtOAuthInstance.UpsertMiner(adminCtx, &UpsertMinerReq{
			User:  user1,
			Miner: mAddr,
		})
		assert.Nil(t, err)
		assert.False(t, ok)

	}
	invalidPermTest := func(ctx context.Context, expect error) {
		mAddr, _ := address.NewFromString(user1Miners[0])
		// Delete miner
		_, err := jwtOAuthInstance.DelMiner(ctx, &DelMinerReq{Miner: mAddr})
		assert.Equal(t, expect, errors.Unwrap(err))
	}
	userCtx := ctxWithUserNameAndAdminPerm(user1)
	validPermTest(userCtx)
	validPermTest(adminCtx)

	invalidPermTest(context.Background(), ErrorUsernameNotFound)
	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)

	// Delete an invalid miner
	mAddr, _ := address.NewFromString(invalidMiner)
	deleted, err := jwtOAuthInstance.DelMiner(adminCtx, &DelMinerReq{Miner: mAddr})
	assert.Nil(t, err)
	assert.False(t, deleted)
}

func addUsersAndSigners(t *testing.T, userSigners map[string][]string) {
	for userName, signers := range userSigners {
		createUserReq := &CreateUserRequest{
			Name:  userName,
			State: 1,
		}

		// with adminCtx admin perm
		// Create users.
		_, err := jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
		assert.NoError(t, err)
		// Add Signer
		signerAddrs := make([]address.Address, 0)
		for _, signer := range signers {
			signerAddr, err := address.NewFromString(signer)
			assert.Nil(t, err)
			signerAddrs = append(signerAddrs, signerAddr)
		}
		err = jwtOAuthInstance.RegisterSigners(adminCtx, &RegisterSignersReq{
			User:    userName,
			Signers: signerAddrs,
		})
		assert.Nil(t, err)

		// with ctx no perm
		// Add Signer
		err = jwtOAuthInstance.RegisterSigners(signCtx, &RegisterSignersReq{
			User:    userName,
			Signers: signerAddrs,
		})
		assert.True(t, errors.Is(err, ErrorPermissionDeny))
	}
}

func testRegisterSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	addUsersAndSigners(t, userSigners)

	// error signer address
	ctx := adminCtx
	_, _ = jwtOAuthInstance.CreateUser(ctx, &CreateUserRequest{
		Name:  "user_01",
		State: 1,
	})
	sAddr, err := address.NewFromString("f0128788")
	assert.Nil(t, err)
	err = jwtOAuthInstance.RegisterSigners(ctx, &RegisterSignersReq{
		User:    "user_01",
		Signers: []address.Address{sAddr},
	})
	assert.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid protocol type")
}

func testSignerExistInUser(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	addUsersAndSigners(t, userSigners)
	// test every user
	for user, signers := range userSigners {

		validPermTest := func(ctx context.Context) {
			for _, signer := range signers {
				sAddr, err := address.NewFromString(signer)
				assert.Nil(t, err)
				bExist, err := jwtOAuthInstance.SignerExistInUser(ctx, &SignerExistInUserReq{
					User:   user,
					Signer: sAddr,
				})
				assert.Nil(t, err)
				assert.True(t, bExist)
			}
		}
		invalidPermTest := func(ctx context.Context, expect error) {
			for _, signer := range signers {
				sAddr, _ := address.NewFromString(signer)
				_, err := jwtOAuthInstance.SignerExistInUser(ctx, &SignerExistInUserReq{
					User:   user,
					Signer: sAddr,
				})
				assert.Equal(t, expect, errors.Unwrap(err))
			}
		}

		userCtx := ctxWithUserNameAndAdminPerm(user)
		validPermTest(userCtx)
		validPermTest(adminCtx)
		invalidPermTest(context.Background(), ErrorUsernameNotFound)
		invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)
	}
}

func testListSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	validUser1 := "test_user_001"
	user1Signers := userSigners[validUser1]

	validPermTest := func(ctx context.Context) {
		// List miners
		resp, err := jwtOAuthInstance.ListSigner(adminCtx, &ListSignerReq{User: validUser1})
		assert.Nil(t, err)
		assert.Equal(t, len(user1Signers), len(resp))
		for _, signer := range resp {
			require.Contains(t, user1Signers, signer.Signer.String())
		}
	}
	invalidPermTest := func(ctx context.Context, expect error) {
		// List miners
		_, err := jwtOAuthInstance.ListSigner(ctx, &ListSignerReq{User: validUser1})
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	userCtx := ctxWithUserNameAndAdminPerm(validUser1)
	validPermTest(userCtx)
	validPermTest(adminCtx)
	invalidPermTest(context.Background(), ErrorUsernameNotFound)
	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)
}

func testHasSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	sAddr, err := address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	assert.Nil(t, err)
	// with adminCtx
	has, err := jwtOAuthInstance.HasSigner(adminCtx, &HasSignerReq{Signer: sAddr})
	assert.Nil(t, err)
	assert.True(t, has)

	sAddr, err = address.NewFromString("f3r72mrymha6wrtb6dzynkzjbnl572az27ddbiq3aovj3d235h2jjgsya4afbf3d37vzfbtsy3dssfnitnhklq")
	assert.Nil(t, err)
	has, err = jwtOAuthInstance.HasSigner(adminCtx, &HasSignerReq{Signer: sAddr})
	assert.Nil(t, err)
	assert.False(t, has)

	// with signCtx
	sAddr, _ = address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	has, err = jwtOAuthInstance.HasSigner(signCtx, &HasSignerReq{Signer: sAddr})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
	assert.False(t, has)
}

func testGetUserBySigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	// with adminCtx
	// Get User By Signer
	sAddr, err := address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	assert.Nil(t, err)
	users, err := jwtOAuthInstance.GetUserBySigner(adminCtx, &GetUserBySignerReq{
		Signer: sAddr,
	})

	names := make([]string, len(users))
	for idx, user := range users {
		names[idx] = user.Name
	}

	assert.Nil(t, err)
	require.Contains(t, names, "test_user_001")
	require.Contains(t, names, "test_user_003")

	// with signCtx
	_, err = jwtOAuthInstance.GetUserBySigner(signCtx, &GetUserBySignerReq{
		Signer: sAddr,
	})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testUnregisterSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	username := "test_user_001"
	sAddr, err := address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	assert.Nil(t, err)

	// with adminCtx admin perm
	adminCtx := adminCtx
	err = jwtOAuthInstance.UnregisterSigners(adminCtx, &UnregisterSignersReq{
		Signers: []address.Address{sAddr},
		User:    username,
	})
	assert.Nil(t, err)

	bExist, err := jwtOAuthInstance.SignerExistInUser(adminCtx, &SignerExistInUserReq{
		Signer: sAddr,
		User:   username,
	})
	assert.Nil(t, err)
	assert.False(t, bExist)

	// with ctx sign perm
	err = jwtOAuthInstance.UnregisterSigners(signCtx, &UnregisterSignersReq{
		Signers: []address.Address{sAddr},
		User:    username,
	})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func testDeleteSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	// Delete signer
	signer, err := address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	assert.Nil(t, err)
	signers := []address.Address{signer}
	userName := "test_user_001"
	validPermTest := func(ctx context.Context) {
		deleted, err := jwtOAuthInstance.DelSigner(ctx, &DelSignerReq{Signer: signer})
		assert.Nil(t, err)
		assert.True(t, deleted)

		// Then get this signer
		has, err := jwtOAuthInstance.HasSigner(adminCtx, &HasSignerReq{Signer: signer})
		assert.Nil(t, err)
		assert.False(t, has)

		// Add Signer back
		err = jwtOAuthInstance.RegisterSigners(adminCtx, &RegisterSignersReq{
			User:    userName,
			Signers: signers,
		})
		assert.Nil(t, err)
	}
	invalidPermTest := func(ctx context.Context, expect error) {
		_, err := jwtOAuthInstance.DelSigner(ctx, &DelSignerReq{Signer: signer})
		assert.Equal(t, expect, errors.Unwrap(err))
	}

	userCtx := ctxWithUserNameAndAdminPerm(userName)
	validPermTest(userCtx)
	validPermTest(adminCtx)
	invalidPermTest(context.Background(), ErrorUsernameNotFound)
	invalidPermTest(signPermAndRandomUsername(), ErrorPermissionDeny)
}

func addUsersAndRateLimits(t *testing.T, userMiners map[string][]string, originLimits []*storage.UserRateLimit) {
	// with adminCtx perm admin
	// Create 3 users and add rate limits
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:  userName,
			State: 0,
		}
		_, _ = jwtOAuthInstance.CreateUser(adminCtx, createUserReq)
	}
	for _, limit := range originLimits {
		id, err := jwtOAuthInstance.UpsertUserRateLimit(adminCtx, &UpsertUserRateLimitReq{
			Id:       limit.Id,
			Name:     limit.Name,
			Service:  limit.Service,
			API:      limit.API,
			ReqLimit: limit.ReqLimit,
		})
		assert.Nil(t, err)
		assert.Equal(t, limit.Id, id)
	}

	// with ctx no perm
	_, err := jwtOAuthInstance.UpsertUserRateLimit(signCtx, &UpsertUserRateLimitReq{
		Id:       originLimits[0].Id,
		Name:     originLimits[0].Name,
		Service:  originLimits[0].Service,
		API:      originLimits[0].API,
		ReqLimit: originLimits[0].ReqLimit,
	})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))

}

func testUpsertUserRateLimit(t *testing.T, userMiners map[string][]string, originLimits []*storage.UserRateLimit) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndRateLimits(t, userMiners, originLimits)
}

func testGetUserRateLimits(t *testing.T, userMiners map[string][]string, originLimits []*storage.UserRateLimit) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndRateLimits(t, userMiners, originLimits)

	// with ctx admin perm
	userName := originLimits[0].Name
	existId := originLimits[0].Id
	resp, err := jwtOAuthInstance.GetUserRateLimits(adminCtx, &GetUserRateLimitsReq{
		Id:   existId,
		Name: userName,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp))
	assert.Equal(t, existId, resp[0].Id)

	// with ctx no perm
	_, err = jwtOAuthInstance.GetUserRateLimits(context.Background(), &GetUserRateLimitsReq{
		Id:   existId,
		Name: userName,
	})
	assert.Equal(t, ErrorPermissionNotFound, errors.Unwrap(err))

	_ = jwtOAuthInstance.DelUserRateLimit(signCtx, &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
}

func testDeleteUserRateLimits(t *testing.T, userMiners map[string][]string, originLimits []*storage.UserRateLimit) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndRateLimits(t, userMiners, originLimits)

	// Test DelUserRateLimit
	userName := originLimits[0].Name
	existId := originLimits[0].Id

	// with ctx admin perm
	err := jwtOAuthInstance.DelUserRateLimit(adminCtx, &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
	assert.Nil(t, err)
	// Try to get it again
	resp, err := jwtOAuthInstance.GetUserRateLimits(adminCtx, &GetUserRateLimitsReq{
		Id:   existId,
		Name: userName,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp))
	// Try to delete again
	err = jwtOAuthInstance.DelUserRateLimit(adminCtx, &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
	assert.NotNil(t, err)

	// with ctx no perm
	err = jwtOAuthInstance.DelUserRateLimit(signCtx, &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
	assert.True(t, errors.Is(err, ErrorPermissionDeny))
}

func TestTokenDecode(t *testing.T) {
	payload := []byte("eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ")
	pb, err := DecodeToBytes(payload)
	if err != nil {
		t.Fatal(err)
	}
	a := map[string]interface{}{}
	err = json.Unmarshal(pb, &a)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["name"], "John Doe")
}

func setup(cfg *config.DBConfig, t *testing.T) {
	var err error
	var dataPath string
	if cfg.Type == "badger" {
		dataPath = t.TempDir()
	}
	theStore, err := storage.NewStore(cfg, dataPath)
	if err != nil {
		t.Fatal(err)
	}

	jwtOAuthInstance = &jwtOAuth{
		store: theStore,
		mp:    newMapper(),
	}
}

func shutdown(cfg *config.DBConfig, t *testing.T) {
	fmt.Printf("shutdown, remove dir:%s\n", cfg.DSN)
	jwtOAuthInstance = nil
	if err := os.RemoveAll(cfg.DSN); err != nil {
		t.Fatal(err)
	}
}

func ctxWithUserNameAndAdminPerm(userName string) context.Context {
	return core.CtxWithName(adminCtx, userName)
}
