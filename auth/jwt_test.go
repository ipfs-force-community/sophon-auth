// stm: #unit
package auth

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/storage"
)

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
	t.Run("test has miner", func(t *testing.T) { testHasMiner(t, userMiners) })
	// stm: @VENUSAUTH_JWT_GET_USER_BY_MINER_001, @VENUSAUTH_JWT_GET_USER_BY_MINER_002, @VENUSAUTH_JWT_GET_USER_BY_MINER_003
	t.Run("test get user by miner", func(t *testing.T) { testGetUserByMiner(t, userMiners) })
	// stm: @VENUSAUTH_JWT_DELETE_MINER_001, @VENUSAUTH_JWT_DELETE_MINER_002
	t.Run("test delete miner", func(t *testing.T) { testDeleteMiner(t, userMiners) })

	// Features about signers
	userSigners := map[string][]string{
		"test_user_001": {"t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha", "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"},
		"test_user_002": {"t3r47fkdzfmtex5ic3jnwlzc7bkpbj7s4d6limyt4f57t3cuqq5nuvhvwv2cu2a6iga2s64vjqcxjqiezyjooq", "t1uqtvvwkkfkkez52ocnqe6vg74qewiwja4t2tiba"},
		"test_user_003": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
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

	token1, err := jwtOAuthInstance.GenerateToken(context.Background(), pl1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(strings.Split(token1, ".")))
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
	ctx := context.Background()
	token1, err := jwtOAuthInstance.GenerateToken(ctx, pl1)
	assert.Nil(t, err)

	// Verify a valid token
	payload1, err := jwtOAuthInstance.Verify(ctx, token1)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(payload1, pl1))

	// Try to verify an invalid token
	invalidToken := "I'm just an invalid token"
	_, err = jwtOAuthInstance.Verify(ctx, invalidToken)
	assert.NotNil(t, err)
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
	ctx := context.Background()
	token1, err := jwtOAuthInstance.GenerateToken(ctx, pl1)
	assert.Nil(t, err)

	// Get token
	tokenInfo1, err := jwtOAuthInstance.GetToken(ctx, token1)
	assert.Nil(t, err)
	assert.Equal(t, pl1.Name, tokenInfo1.Name)
	assert.Equal(t, pl1.Perm, tokenInfo1.Perm)
	// Try to get invalid token
	invalidToken := "I'm just an invalid token"
	_, err = jwtOAuthInstance.GetToken(ctx, invalidToken)
	assert.NotNil(t, err)
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
	ctx := context.Background()
	token1, err := jwtOAuthInstance.GenerateToken(ctx, pl1)
	assert.Nil(t, err)

	// Get token by name
	tokenInfoList1, err := jwtOAuthInstance.GetTokenByName(ctx, "test-token-01")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tokenInfoList1))
	assert.Equal(t, token1, tokenInfoList1[0].Token)
	// Try to get token by wrong name
	tokenInfoInvalid, err := jwtOAuthInstance.GetTokenByName(ctx, "invalid_name")
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
	ctx := context.Background()
	_, err := jwtOAuthInstance.GenerateToken(ctx, pl1)
	assert.Nil(t, err)
	_, err = jwtOAuthInstance.GenerateToken(ctx, pl2)
	assert.Nil(t, err)

	allTokenInfos, err := jwtOAuthInstance.Tokens(ctx, 0, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allTokenInfos))
	// with skip or limit
	allTokenInfos, err = jwtOAuthInstance.Tokens(ctx, 1, 10)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))

	allTokenInfos, err = jwtOAuthInstance.Tokens(ctx, 0, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))

	allTokenInfos, err = jwtOAuthInstance.Tokens(ctx, 2, 10)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(allTokenInfos))
}

func testRemoveAndRecoverToken(t *testing.T) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	// First generate a token
	pl1 := &JWTPayload{
		Name:  "test-token-01",
		Perm:  "admin",
		Extra: "",
	}
	ctx := context.Background()
	token1, err := jwtOAuthInstance.GenerateToken(ctx, pl1)
	assert.Nil(t, err)

	// token is usable.
	err = jwtOAuthInstance.RecoverToken(ctx, token1)
	assert.Error(t, err)

	// Remove a token
	err = jwtOAuthInstance.RemoveToken(ctx, token1)
	assert.Nil(t, err)

	_, err = jwtOAuthInstance.Verify(ctx, token1)
	assert.NotNil(t, err)

	tokenInfoList1, err := jwtOAuthInstance.GetTokenByName(ctx, "test-token-01")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tokenInfoList1))

	// Recover a token
	err = jwtOAuthInstance.RecoverToken(ctx, token1)
	assert.Nil(t, err)
	payload1, err := jwtOAuthInstance.Verify(ctx, token1)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(payload1, pl1))
	allTokenInfos, err := jwtOAuthInstance.Tokens(ctx, 0, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))
}

func createUsers(t *testing.T, userMiners map[string][]string) {
	ctx := context.Background()
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

	ctx := context.Background()

	existUserName := "test_user_001"
	comment := "test comment"
	// Create 3 users
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:    userName,
			Comment: &comment,
			State:   0,
		}
		resp, err := jwtOAuthInstance.CreateUser(ctx, createUserReq)
		assert.Nil(t, err)
		assert.Equal(t, userName, resp.Name)
		assert.Equal(t, "test comment", resp.Comment)
	}
	// Create duplicate user
	_, err := jwtOAuthInstance.CreateUser(ctx, &CreateUserRequest{Name: existUserName})
	assert.NotNil(t, err)
}

func testGetUser(t *testing.T, userMiners map[string][]string) {
	existUserName := "test_user_001"
	invalidUserName := "invalid_name"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	ctx := context.Background()
	// HasUser
	exist, err := jwtOAuthInstance.HasUser(ctx, &HasUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.True(t, exist)
	exist, err = jwtOAuthInstance.HasUser(ctx, &HasUserRequest{Name: invalidUserName})
	assert.Nil(t, err)
	assert.False(t, exist)

	u, err := jwtOAuthInstance.GetUser(ctx, &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, u.Name, existUserName)
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

	ctx := context.Background()
	err := jwtOAuthInstance.VerifyUsers(ctx, &VerifyUsersReq{Names: usernames})
	assert.Nil(t, err)
}

func testListUser(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	allUserInfos, err := jwtOAuthInstance.ListUsers(context.Background(), &ListUsersRequest{
		Page:  &core.Page{},
		State: int(core.UserStateUndefined),
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allUserInfos))
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
	ctx := context.Background()
	err := jwtOAuthInstance.UpdateUser(ctx, updateUserReq)
	assert.Nil(t, err)
	// Then get this user
	outPutUser1, err := jwtOAuthInstance.GetUser(ctx, &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, "New Comment", outPutUser1.Comment)

	// invalid user name
	err = jwtOAuthInstance.UpdateUser(ctx, &UpdateUserRequest{})
	assert.Error(t, err)
}

func testDeleteAndRecoverUser(t *testing.T, userMiners map[string][]string) {
	existUserName := "test_user_001"
	invalidUserName := "invalid_name"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	ginCtx := &gin.Context{}

	// Delete User
	err := jwtOAuthInstance.DeleteUser(ginCtx, &DeleteUserRequest{Name: existUserName})
	assert.Nil(t, err)
	// Then try to get this user
	_, err = jwtOAuthInstance.GetUser(ginCtx, &GetUserRequest{Name: existUserName})
	assert.NotNil(t, err)
	// And also list users now
	allUserInfos, err := jwtOAuthInstance.ListUsers(ginCtx, &ListUsersRequest{
		Page:  &core.Page{},
		State: int(core.UserStateUndefined),
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allUserInfos))

	// Try to delete non-existing users
	err = jwtOAuthInstance.DeleteUser(ginCtx, &DeleteUserRequest{Name: invalidUserName})
	assert.NotNil(t, err)

	// Recover user
	err = jwtOAuthInstance.RecoverUser(ginCtx, &RecoverUserRequest{Name: existUserName})
	assert.Nil(t, err)
	// Then get this user
	outPutUser1, err := jwtOAuthInstance.GetUser(ginCtx, &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, existUserName, outPutUser1.Name)

	// Try to recover an invalid user
	err = jwtOAuthInstance.RecoverUser(ginCtx, &RecoverUserRequest{Name: invalidUserName})
	assert.NotNil(t, err)

	// Try to recover a valid, but not deleted user
	err = jwtOAuthInstance.RecoverUser(ginCtx, &RecoverUserRequest{Name: existUserName})
	assert.NotNil(t, err)
}

func addUsersAndMiners(t *testing.T, userMiners map[string][]string) {
	ctx := context.Background()
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
			ifCreate, err := jwtOAuthInstance.UpsertMiner(ctx, &UpsertMinerReq{
				User:       userName,
				Miner:      minerID,
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

	ctx := context.Background()
	// error signer address
	_, _ = jwtOAuthInstance.CreateUser(ctx, &CreateUserRequest{
		Name:  "user_01",
		State: 1,
	})
	isCreate, err := jwtOAuthInstance.UpsertMiner(ctx, &UpsertMinerReq{User: "user_01", Miner: "f01034"})
	assert.Nil(t, err)
	assert.True(t, isCreate)

	_, err = jwtOAuthInstance.UpsertMiner(ctx, &UpsertMinerReq{
		User:  "user_01",
		Miner: "f1mpvdqt2acgihevibd4greavlsfn3dfph5sckc2a",
	})
	assert.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid protocol type")
}

func testListMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	validUser1 := "test_user_001"
	user1Miners := []string{"t01000", "t01002", "t01003"}
	ctx := context.Background()
	// List miners
	resp, err := jwtOAuthInstance.ListMiners(ctx, &ListMinerReq{User: validUser1})
	assert.Nil(t, err)
	assert.Equal(t, len(user1Miners), len(resp))
	sort.Slice(resp, func(i, j int) bool { return resp[i].Miner < resp[j].Miner })
	for i := 0; i < len(user1Miners); i++ {
		assert.Equal(t, user1Miners[i], resp[i].Miner)
		assert.Equal(t, validUser1, resp[i].User)
		assert.Equal(t, true, resp[i].OpenMining)
	}
}

func testHasMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	ctx := context.Background()

	// Has Miner
	has, err := jwtOAuthInstance.HasMiner(ctx, &HasMinerRequest{Miner: "t01000"})
	assert.Nil(t, err)
	assert.True(t, has)

	// Miner Exist In Account
	exist, err := jwtOAuthInstance.MinerExistInUser(ctx, &MinerExistInUserRequest{Miner: "t01000", User: "test_user_001"})
	assert.Nil(t, err)
	assert.True(t, exist)

	exist, err = jwtOAuthInstance.MinerExistInUser(ctx, &MinerExistInUserRequest{Miner: "t01000", User: "test_user_002"})
	assert.Nil(t, err)
	assert.False(t, exist)

	_, err = jwtOAuthInstance.HasMiner(ctx, &HasMinerRequest{Miner: "invalid address"})
	assert.Error(t, err)
}

func testGetUserByMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	ctx := context.Background()
	// Get User By Miner
	validUser1 := "test_user_001"
	user1Miners := []string{"t01000", "t01002", "t01003"}
	userInfo, err := jwtOAuthInstance.GetUserByMiner(ctx, &GetUserByMinerRequest{
		Miner: user1Miners[1],
	})
	assert.Nil(t, err)
	assert.Equal(t, validUser1, userInfo.Name)

	// invalid miner address
	_, err = jwtOAuthInstance.GetUserByMiner(ctx, &GetUserByMinerRequest{
		Miner: "invalid address",
	})
	assert.Error(t, err)

	// miner address not exist
	_, err = jwtOAuthInstance.GetUserByMiner(ctx, &GetUserByMinerRequest{
		Miner: "f01989787",
	})
	assert.Error(t, err)
}

func testDeleteMiner(t *testing.T, userMiners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	user1Miners := []string{"t01000", "t01002", "t01003"}
	invalidMiner := "t02000"
	ctx := context.Background()
	// Delete miner
	deleted, err := jwtOAuthInstance.DelMiner(ctx, &DelMinerReq{Miner: user1Miners[0]})
	assert.Nil(t, err)
	assert.True(t, deleted)
	// Then get this miner
	has, err := jwtOAuthInstance.HasMiner(ctx, &HasMinerRequest{Miner: user1Miners[0]})
	assert.Nil(t, err)
	assert.False(t, has)
	// Try to get user by this miner
	_, err = jwtOAuthInstance.GetUserByMiner(ctx, &GetUserByMinerRequest{
		Miner: user1Miners[0],
	})
	assert.NotNil(t, err)

	// Delete an invalid miner
	deleted, err = jwtOAuthInstance.DelMiner(ctx, &DelMinerReq{Miner: invalidMiner})
	assert.Nil(t, err)
	assert.False(t, deleted)
}

func addUsersAndSigners(t *testing.T, userSigners map[string][]string) {
	for userName, signers := range userSigners {
		createUserReq := &CreateUserRequest{
			Name:  userName,
			State: 1,
		}

		ctx := context.Background()
		// Create users.
		_, _ = jwtOAuthInstance.CreateUser(ctx, createUserReq)
		// Add Signer
		err := jwtOAuthInstance.RegisterSigners(ctx, &RegisterSignersReq{
			User:    userName,
			Signers: signers,
		})
		assert.Nil(t, err)
	}
}

func testRegisterSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	addUsersAndSigners(t, userSigners)

	// error signer address
	ctx := context.Background()
	_, _ = jwtOAuthInstance.CreateUser(ctx, &CreateUserRequest{
		Name:  "user_01",
		State: 1,
	})
	err := jwtOAuthInstance.RegisterSigners(ctx, &RegisterSignersReq{
		User:    "user_01",
		Signers: []string{"f0128788"},
	})
	assert.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid protocol type")

	err = jwtOAuthInstance.RegisterSigners(ctx, &RegisterSignersReq{
		User:    "user_01",
		Signers: []string{"128788"},
	})
	assert.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid signer address")
}

func testSignerExistInUser(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	addUsersAndSigners(t, userSigners)

	ctx := context.Background()
	for user, signers := range userSigners {
		for _, signer := range signers {
			bExist, err := jwtOAuthInstance.SignerExistInUser(ctx, &SignerExistInUserReq{
				User:   user,
				Signer: signer,
			})
			assert.Nil(t, err)
			assert.True(t, bExist)
		}
	}
}

func testListSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	validUser1 := "test_user_001"
	user1Signers := userSigners[validUser1]
	// List miners
	resp, err := jwtOAuthInstance.ListSigner(context.Background(), &ListSignerReq{User: validUser1})
	assert.Nil(t, err)
	assert.Equal(t, len(user1Signers), len(resp))
	for _, signer := range resp {
		require.Contains(t, user1Signers, signer.Signer)
	}
}

func testHasSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)
	ctx := context.Background()

	has, err := jwtOAuthInstance.HasSigner(ctx, &HasSignerReq{Signer: "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"})
	assert.Nil(t, err)
	assert.True(t, has)

	has, err = jwtOAuthInstance.HasSigner(ctx, &HasSignerReq{Signer: "f3r72mrymha6wrtb6dzynkzjbnl572az27ddbiq3aovj3d235h2jjgsya4afbf3d37vzfbtsy3dssfnitnhklq"})
	assert.Nil(t, err)
	assert.False(t, has)
}

func testGetUserBySigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	// Get User By Signer
	signer := "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"
	users, err := jwtOAuthInstance.GetUserBySigner(context.Background(), &GetUserBySignerReq{
		Signer: signer,
	})

	names := make([]string, len(users))
	for idx, user := range users {
		names[idx] = user.Name
	}

	assert.Nil(t, err)
	require.Contains(t, names, "test_user_001")
	require.Contains(t, names, "test_user_003")
}

func testUnregisterSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	username := "test_user_001"
	signer := "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"
	ctx := context.Background()
	err := jwtOAuthInstance.UnregisterSigners(ctx, &UnregisterSignersReq{
		Signers: []string{signer},
		User:    username,
	})

	assert.Nil(t, err)

	bExist, err := jwtOAuthInstance.SignerExistInUser(ctx, &SignerExistInUserReq{
		Signer: signer,
		User:   username,
	})
	assert.Nil(t, err)
	assert.False(t, bExist)
}

func testDeleteSigner(t *testing.T, userSigners map[string][]string) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndSigners(t, userSigners)

	// Delete signer
	signer := "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"
	ctx := context.Background()
	deleted, err := jwtOAuthInstance.DelSigner(ctx, &DelSignerReq{Signer: signer})
	assert.Nil(t, err)
	assert.True(t, deleted)

	// Then get this signer
	has, err := jwtOAuthInstance.HasSigner(ctx, &HasSignerReq{Signer: signer})
	assert.Nil(t, err)
	assert.False(t, has)
}

func addUsersAndRateLimits(t *testing.T, userMiners map[string][]string, originLimits []*storage.UserRateLimit) {
	ctx := context.Background()
	// Create 3 users and add rate limits
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:  userName,
			State: 0,
		}
		_, _ = jwtOAuthInstance.CreateUser(ctx, createUserReq)
	}
	for _, limit := range originLimits {
		id, err := jwtOAuthInstance.UpsertUserRateLimit(ctx, &UpsertUserRateLimitReq{
			Id:       limit.Id,
			Name:     limit.Name,
			Service:  limit.Service,
			API:      limit.API,
			ReqLimit: limit.ReqLimit,
		})
		assert.Nil(t, err)
		assert.Equal(t, limit.Id, id)
	}
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

	// Test GetUserRateLimits
	userName := originLimits[0].Name
	existId := originLimits[0].Id
	resp, err := jwtOAuthInstance.GetUserRateLimits(context.Background(), &GetUserRateLimitsReq{
		Id:   existId,
		Name: userName,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp))
	assert.Equal(t, existId, resp[0].Id)
}

func testDeleteUserRateLimits(t *testing.T, userMiners map[string][]string, originLimits []*storage.UserRateLimit) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndRateLimits(t, userMiners, originLimits)

	// Test DelUserRateLimit
	userName := originLimits[0].Name
	existId := originLimits[0].Id

	ctx := context.Background()

	err := jwtOAuthInstance.DelUserRateLimit(ctx, &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
	assert.Nil(t, err)
	// Try to get it again
	resp, err := jwtOAuthInstance.GetUserRateLimits(ctx, &GetUserRateLimitsReq{
		Id:   existId,
		Name: userName,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp))
	// Try to delete again
	err = jwtOAuthInstance.DelUserRateLimit(ctx, &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
	assert.NotNil(t, err)
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
	if cfg.Type == "badger" {
		if cfg.DSN, err = ioutil.TempDir("", "auth-datastore"); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("tmp badger store : %s\n", cfg.DSN)
	}
	theStore, err := storage.NewStore(cfg, cfg.DSN)
	if err != nil {
		t.Fatal(err)
	}

	secret, err := config.RandSecret()
	if err != nil {
		t.Fatal(err)
	}
	sec, err := hex.DecodeString(hex.EncodeToString(secret))
	if err != nil {
		t.Fatal(err)
	}
	jwtOAuthInstance = &jwtOAuth{
		secret: jwt.NewHS256(sec),
		store:  theStore,
		mp:     newMapper(),
	}
}

func shutdown(cfg *config.DBConfig, t *testing.T) {
	fmt.Printf("shutdown, remove dir:%s\n", cfg.DSN)
	jwtOAuthInstance = nil
	if err := os.RemoveAll(cfg.DSN); err != nil {
		t.Fatal(err)
	}
}
