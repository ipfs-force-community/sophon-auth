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

	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/gbrlsnchs/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJwt(t *testing.T) {
	var limitStrs = `[{"Id":"794fc9a4-2b80-4503-835a-7e8e27360b3d","Name":"test_user_01","Service":"","API":"","ReqLimit":{"Cap":10,"ResetDur":120000000000}},{"Id":"252f581e-cbd2-4a61-a517-0b7df65013aa","Name":"test_user_02","Service":"","API":"","ReqLimit":{"Cap":10,"ResetDur":72000000000000}}]`
	var originLimits []*storage.UserRateLimit

	var userMiners = map[string]map[string]interface{}{
		"test_user_001": {"t01000": nil, "t01002": nil, "t01003": nil},
		"test_user_002": {"t01004": nil, "t01005": nil, "t01006": nil},
		"test_user_003": {"t01007": nil, "t01008": nil, "t01009": nil},
	}

	if err := json.Unmarshal([]byte(limitStrs), &originLimits); err != nil {
		t.Fatalf("initialize origin Ratelimit failed:%s", err.Error())
	}

	// Features about tokens
	t.Run("generate token", testGenerateToken)
	t.Run("verify token", testVerifyToken)
	t.Run("get token", testGetToken)
	t.Run("get token by name", testGetTokenByName)
	t.Run("list all tokens", testTokenList)
	t.Run("remove and recover tokens", testRemoveAndRecoverToken)

	// Features about users
	t.Run("test create user", func(t *testing.T) { testCreateUser(t, userMiners) })
	t.Run("test get user", func(t *testing.T) { testGetUser(t, userMiners) })
	t.Run("test list user", func(t *testing.T) { testListUser(t, userMiners) })
	t.Run("test update user", func(t *testing.T) { testUpdateUser(t, userMiners) })
	t.Run("test delete and recover user", func(t *testing.T) { testDeleteAndRecoverUser(t, userMiners) })

	// Features about miners
	t.Run("test upsert miner", func(t *testing.T) { testUpsertMiner(t, userMiners) })
	t.Run("test list miner", func(t *testing.T) { testListMiner(t, userMiners) })
	t.Run("test has miner", func(t *testing.T) { testHasMiner(t, userMiners) })
	t.Run("test get user by miner", func(t *testing.T) { testGetUserByMiner(t, userMiners) })
	t.Run("test delete miner", func(t *testing.T) { testDeleteMiner(t, userMiners) })

	// Features about rate limits
	t.Run("test upsert rate limit", func(t *testing.T) { testUpsertUserRateLimit(t, userMiners, originLimits) })
	t.Run("test get rate limit", func(t *testing.T) { testGetUserRateLimits(t, userMiners, originLimits) })
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
	token1, err := jwtOAuthInstance.GenerateToken(context.Background(), pl1)
	assert.Nil(t, err)

	// Verify a valid token
	payload1, err := jwtOAuthInstance.Verify(context.Background(), token1)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(payload1, pl1))

	// Try to verify an invalid token
	invalidToken := "I'm just an invalid token"
	_, err = jwtOAuthInstance.Verify(context.Background(), invalidToken)
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
	token1, err := jwtOAuthInstance.GenerateToken(context.Background(), pl1)
	assert.Nil(t, err)

	// Get token
	tokenInfo1, err := jwtOAuthInstance.GetToken(context.Background(), token1)
	assert.Nil(t, err)
	assert.Equal(t, pl1.Name, tokenInfo1.Name)
	assert.Equal(t, pl1.Perm, tokenInfo1.Perm)
	// Try to get invalid token
	invalidToken := "I'm just an invalid token"
	_, err = jwtOAuthInstance.GetToken(context.Background(), invalidToken)
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
	token1, err := jwtOAuthInstance.GenerateToken(context.Background(), pl1)
	assert.Nil(t, err)

	// Get token by name
	tokenInfoList1, err := jwtOAuthInstance.GetTokenByName(context.Background(), "test-token-01")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tokenInfoList1))
	assert.Equal(t, token1, tokenInfoList1[0].Token)
	// Try to get token by wrong name
	tokenInfoInvalid, err := jwtOAuthInstance.GetTokenByName(context.Background(), "invalid_name")
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
	_, err := jwtOAuthInstance.GenerateToken(context.Background(), pl1)
	assert.Nil(t, err)
	_, err = jwtOAuthInstance.GenerateToken(context.Background(), pl2)
	assert.Nil(t, err)

	allTokenInfos, err := jwtOAuthInstance.Tokens(context.Background(), 0, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allTokenInfos))
	// with skip or limit
	allTokenInfos, err = jwtOAuthInstance.Tokens(context.Background(), 1, 10)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))

	allTokenInfos, err = jwtOAuthInstance.Tokens(context.Background(), 0, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))

	allTokenInfos, err = jwtOAuthInstance.Tokens(context.Background(), 2, 10)
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
	token1, err := jwtOAuthInstance.GenerateToken(context.Background(), pl1)
	assert.Nil(t, err)

	// Remove a token
	err = jwtOAuthInstance.RemoveToken(context.Background(), token1)
	assert.Nil(t, err)

	_, err = jwtOAuthInstance.Verify(context.Background(), token1)
	assert.NotNil(t, err)

	tokenInfoList1, err := jwtOAuthInstance.GetTokenByName(context.Background(), "test-token-01")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tokenInfoList1))

	// Recover a token
	err = jwtOAuthInstance.RecoverToken(context.Background(), token1)
	assert.Nil(t, err)
	payload1, err := jwtOAuthInstance.Verify(context.Background(), token1)
	assert.Nil(t, err)
	assert.True(t, reflect.DeepEqual(payload1, pl1))
	allTokenInfos, err := jwtOAuthInstance.Tokens(context.Background(), 0, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(allTokenInfos))
}

func createUsers(t *testing.T, userMiners map[string]map[string]interface{}) {
	// Create 3 users
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:       userName,
			Comment:    "",
			State:      0,
			SourceType: 0,
		}
		resp, err := jwtOAuthInstance.CreateUser(context.Background(), createUserReq)
		assert.Nil(t, err)
		assert.Equal(t, userName, resp.Name)
		assert.Equal(t, "", resp.Comment)
	}
}

func testCreateUser(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)

	existUserName := "test_user_001"
	// Create 3 users
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:       userName,
			Comment:    "",
			State:      0,
			SourceType: 0,
		}
		resp, err := jwtOAuthInstance.CreateUser(context.Background(), createUserReq)
		assert.Nil(t, err)
		assert.Equal(t, userName, resp.Name)
		assert.Equal(t, "", resp.Comment)
	}
	// Create duplicate user
	_, err := jwtOAuthInstance.CreateUser(context.Background(), &CreateUserRequest{Name: existUserName})
	assert.NotNil(t, err)
}

func testGetUser(t *testing.T, userMiners map[string]map[string]interface{}) {
	existUserName := "test_user_001"
	invalidUserName := "invalid_name"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)
	// HasUser
	exist, err := jwtOAuthInstance.HasUser(context.Background(), &HasUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.True(t, exist)
	exist, err = jwtOAuthInstance.HasUser(context.Background(), &HasUserRequest{Name: invalidUserName})
	assert.Nil(t, err)
	assert.False(t, exist)
}

func testListUser(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	allUserInfos, err := jwtOAuthInstance.ListUsers(context.Background(), &ListUsersRequest{
		Page: &core.Page{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allUserInfos))
}

func testUpdateUser(t *testing.T, userMiners map[string]map[string]interface{}) {
	existUserName := "test_user_001"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	// Update a user
	updateUserReq := &UpdateUserRequest{
		KeySum:  2,
		Name:    existUserName,
		Comment: "New Comment",
	}
	err := jwtOAuthInstance.UpdateUser(context.Background(), updateUserReq)
	assert.Nil(t, err)
	// Then get this user
	outPutUser1, err := jwtOAuthInstance.GetUser(context.Background(), &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, "New Comment", outPutUser1.Comment)
}

func testDeleteAndRecoverUser(t *testing.T, userMiners map[string]map[string]interface{}) {
	existUserName := "test_user_001"
	invalidUserName := "invalid_name"

	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	createUsers(t, userMiners)

	// Delete User
	err := jwtOAuthInstance.DeleteUser(&gin.Context{}, &DeleteUserRequest{Name: existUserName})
	assert.Nil(t, err)
	// Then try to get this user
	_, err = jwtOAuthInstance.GetUser(context.Background(), &GetUserRequest{Name: existUserName})
	assert.NotNil(t, err)
	// And also list users now
	allUserInfos, err := jwtOAuthInstance.ListUsers(context.Background(), &ListUsersRequest{
		Page: &core.Page{},
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(allUserInfos))

	// Try to delete non-existing users
	err = jwtOAuthInstance.DeleteUser(&gin.Context{}, &DeleteUserRequest{Name: invalidUserName})
	assert.NotNil(t, err)

	// Recover user
	err = jwtOAuthInstance.RecoverUser(&gin.Context{}, &RecoverUserRequest{Name: existUserName})
	assert.Nil(t, err)
	// Then get this user
	outPutUser1, err := jwtOAuthInstance.GetUser(context.Background(), &GetUserRequest{Name: existUserName})
	assert.Nil(t, err)
	assert.Equal(t, existUserName, outPutUser1.Name)

	// Try to recover an invalid user
	err = jwtOAuthInstance.RecoverUser(&gin.Context{}, &RecoverUserRequest{Name: invalidUserName})
	assert.NotNil(t, err)

	// Try to recover a valid, but not deleted user
	err = jwtOAuthInstance.RecoverUser(&gin.Context{}, &RecoverUserRequest{Name: existUserName})
	assert.NotNil(t, err)
}

func addUsersAndMiners(t *testing.T, userMiners map[string]map[string]interface{}) {
	// Create 3 users and add miners
	for userName, minerMap := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:       userName,
			Comment:    "",
			State:      0,
			SourceType: 0,
		}
		// Create users.
		_, _ = jwtOAuthInstance.CreateUser(context.Background(), createUserReq)
		// Add miners
		for minerName := range minerMap {
			ifCreate, err := jwtOAuthInstance.UpsertMiner(context.Background(), &UpsertMinerReq{
				User:       userName,
				Miner:      minerName,
				OpenMining: true,
			})
			assert.Nil(t, err)
			assert.True(t, ifCreate)
		}
	}
}

func testUpsertMiner(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)
}

func testListMiner(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	validUser1 := "test_user_001"
	user1Miners := []string{"t01000", "t01002", "t01003"}
	// List miners
	resp, err := jwtOAuthInstance.ListMiners(context.Background(), &ListMinerReq{User: validUser1})
	assert.Nil(t, err)
	assert.Equal(t, len(user1Miners), len(resp))
	sort.Slice(resp, func(i, j int) bool { return resp[i].Miner < resp[j].Miner })
	for i := 0; i < len(user1Miners); i++ {
		assert.Equal(t, user1Miners[i], resp[i].Miner)
		assert.Equal(t, validUser1, resp[i].User)
		assert.Equal(t, true, resp[i].OpenMining)
	}
}

func testHasMiner(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	// Has Miner
	user1Miners := []string{"t01000", "t01002", "t01003"}
	has, err := jwtOAuthInstance.HasMiner(context.Background(), &HasMinerRequest{Miner: user1Miners[0]})
	assert.Nil(t, err)
	assert.True(t, has)
}

func testGetUserByMiner(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	// Get User By Miner
	validUser1 := "test_user_001"
	user1Miners := []string{"t01000", "t01002", "t01003"}
	userInfo, err := jwtOAuthInstance.GetUserByMiner(context.Background(), &GetUserByMinerRequest{
		Miner: user1Miners[1],
	})
	assert.Nil(t, err)
	assert.Equal(t, validUser1, userInfo.Name)
}

func testDeleteMiner(t *testing.T, userMiners map[string]map[string]interface{}) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndMiners(t, userMiners)

	user1Miners := []string{"t01000", "t01002", "t01003"}
	invalidMiner := "t02000"
	// Delete miner
	deleted, err := jwtOAuthInstance.DelMiner(context.Background(), &DelMinerReq{Miner: user1Miners[0]})
	assert.Nil(t, err)
	assert.True(t, deleted)
	// Then get this miner
	has, err := jwtOAuthInstance.HasMiner(context.Background(), &HasMinerRequest{Miner: user1Miners[0]})
	assert.Nil(t, err)
	assert.False(t, has)
	// Try to get user by this miner
	_, err = jwtOAuthInstance.GetUserByMiner(context.Background(), &GetUserByMinerRequest{
		Miner: user1Miners[0],
	})
	assert.NotNil(t, err)

	// Delete an invalid miner
	deleted, err = jwtOAuthInstance.DelMiner(context.Background(), &DelMinerReq{Miner: invalidMiner})
	assert.Nil(t, err)
	assert.False(t, deleted)
}

func addUsersAndRateLimits(t *testing.T, userMiners map[string]map[string]interface{}, originLimits []*storage.UserRateLimit) {
	// Create 3 users and add rate limits
	for userName := range userMiners {
		createUserReq := &CreateUserRequest{
			Name:       userName,
			Comment:    "",
			State:      0,
			SourceType: 0,
		}
		_, _ = jwtOAuthInstance.CreateUser(context.Background(), createUserReq)
	}
	for _, limit := range originLimits {
		id, err := jwtOAuthInstance.UpsertUserRateLimit(context.Background(), &UpsertUserRateLimitReq{
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

func testUpsertUserRateLimit(t *testing.T, userMiners map[string]map[string]interface{}, originLimits []*storage.UserRateLimit) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndRateLimits(t, userMiners, originLimits)
}

func testGetUserRateLimits(t *testing.T, userMiners map[string]map[string]interface{}, originLimits []*storage.UserRateLimit) {
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

func testDeleteUserRateLimits(t *testing.T, userMiners map[string]map[string]interface{}, originLimits []*storage.UserRateLimit) {
	cfg := config.DBConfig{Type: "badger"}
	setup(&cfg, t)
	defer shutdown(&cfg, t)
	addUsersAndRateLimits(t, userMiners, originLimits)

	// Test DelUserRateLimit
	userName := originLimits[0].Name
	existId := originLimits[0].Id

	err := jwtOAuthInstance.DelUserRateLimit(context.Background(), &DelUserRateLimitReq{
		Name: userName,
		Id:   existId,
	})
	assert.Nil(t, err)
	// Try to get it again
	resp, err := jwtOAuthInstance.GetUserRateLimits(context.Background(), &GetUserRateLimitsReq{
		Id:   existId,
		Name: userName,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp))
	// Try to delete again
	err = jwtOAuthInstance.DelUserRateLimit(context.Background(), &DelUserRateLimitReq{
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
