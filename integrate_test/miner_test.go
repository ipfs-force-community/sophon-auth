// stm: #integration
package integrate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
)

func TestMinerApis(t *testing.T) {
	// stm: @VENUSAUTH_APP_UPSERT_MINER_001, @VENUSAUTH_APP_UPSERT_MINER_002, @VENUSAUTH_APP_UPSERT_MINER_003
	t.Run("upsert miner", testUpsertMiners)
	// stm: @VENUSAUTH_APP_LIST_MINERS_001, @VENUSAUTH_APP_LIST_MINERS_002
	t.Run("list miner by user", testListMinerByUser)
	// stm: @VENUSAUTH_APP_HAS_MINER_001, @VENUSAUTH_APP_HAS_MINER_002, @VENUSAUTH_APP_HAS_MINER_003
	t.Run("has miner", testHasMiner)
	t.Run("miner exist in user", testMinerExistInUser)
	// stm: @VENUSAUTH_APP_GET_USERS_BY_MINER_001, @VENUSAUTH_APP_GET_USERS_BY_MINER_002, @VENUSAUTH_APP_GET_USERS_BY_MINER_003
	t.Run("get user by miner", testGetUserByMiner)
	// stm: @VENUSAUTH_APP_DEL_MINER_001, @VENUSAUTH_APP_DEL_MINER_003
	t.Run("delete miner", testDeleteMiner)
}

func setupAndAddMiners(t *testing.T) (*jwtclient.AuthClient, *auth.OutputUser, string) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	userName := "Rennbon"
	miner1 := "t01000"
	miner2 := "t01002"

	// Create a user
	user, err := client.CreateUser(&auth.CreateUserRequest{Name: userName})
	assert.Nil(t, err)
	// Add 2 miners
	success, err := client.UpsertMiner(userName, miner1, true)
	assert.Nil(t, err)
	assert.True(t, success)
	success, err = client.UpsertMiner(userName, miner2, true)
	assert.Nil(t, err)
	assert.True(t, success)

	user.Miners = append(user.Miners, &auth.OutputMiner{Miner: miner1, User: userName},
		&auth.OutputMiner{Miner: miner2, User: userName})

	return client, user, tmpDir
}

func testUpsertMiners(t *testing.T) {
	c, user, tmpDir := setupAndAddMiners(t)

	// `ShouldBind` failed
	_, err := c.UpsertMiner("", "f01034", true)
	assert.Error(t, err)

	// invalid address error
	_, err = c.UpsertMiner(user.Name, address.Undef.String(), true)
	assert.Error(t, err)

	shutdown(t, tmpDir)
}

func testListMinerByUser(t *testing.T) {
	client, user, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	// List miner by user
	listResp, err := client.ListMiners(user.Name)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(listResp))

	// `ShouldBind` failed
	_, err = client.ListMiners("")
	assert.Error(t, err)
}

func testHasMiner(t *testing.T) {
	client, _, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	miner2 := "t01002"
	miner3 := "t01004"

	// Has miner
	has, err := client.HasMiner(&auth.HasMinerRequest{Miner: miner2})
	assert.Nil(t, err)
	assert.True(t, has)

	// Has invalid miner
	has, err = client.HasMiner(&auth.HasMinerRequest{Miner: miner3})
	assert.Nil(t, err)
	assert.False(t, has)

	// `ShouldBind` failed
	has, err = client.HasMiner(&auth.HasMinerRequest{})
	assert.Error(t, err)
	assert.False(t, has)
}

func testMinerExistInUser(t *testing.T) {
	client, user, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	notExistMiner := "t010010"

	exist, err := client.MinerExistInUser(user.Name, user.Miners[0].Miner)
	assert.Nil(t, err)
	assert.True(t, exist)

	exist, err = client.MinerExistInUser(user.Name, notExistMiner)
	assert.Nil(t, err)
	assert.False(t, exist)
}

func testGetUserByMiner(t *testing.T) {
	client, user, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	// Get user by miner
	getUserInfo, err := client.GetUserByMiner(&auth.GetUserByMinerRequest{Miner: user.Miners[0].Miner})
	assert.Nil(t, err)
	assert.Equal(t, user.Name, getUserInfo.Name)

	// `ShouldBind` failed
	_, err = client.GetUserByMiner(&auth.GetUserByMinerRequest{})
	assert.Error(t, err)

	// miner not exists error
	_, err = client.GetUserByMiner(&auth.GetUserByMinerRequest{Miner: "f011112222233333"})
	assert.Error(t, err)
}

func testDeleteMiner(t *testing.T) {
	client, user, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	notExistMiner := "t01004"
	// Delete a miner
	success, err := client.DelMiner(user.Miners[0].Miner)
	assert.Nil(t, err)
	assert.True(t, success)

	// Check this miner
	has, err := client.HasMiner(&auth.HasMinerRequest{Miner: user.Miners[0].Miner})
	assert.Nil(t, err)
	assert.False(t, has)

	// Try to delete not exist miner
	success, err = client.DelMiner(notExistMiner)
	assert.Nil(t, err)
	assert.False(t, success)

	// Try to delete a invalid miner
	_, err = client.DelMiner("abcdfghijk")
	assert.Error(t, err)
}
