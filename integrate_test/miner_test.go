package integrate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
)

func TestMinerApis(t *testing.T) {
	t.Run("upsert miner", testUpsertMiners)
	t.Run("list miner by user", testListMinerByUser)
	t.Run("has miner", testHasMiner)
	t.Run("miner exist in user", testMinerExistInUser)
	t.Run("get user by miner", testGetUserByMiner)
	t.Run("delete miner", testDeleteMiner)
}

func setupAndAddMiners(t *testing.T) (*jwtclient.AuthClient, string) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	userName := "Rennbon"
	miner1 := "t01000"
	miner2 := "t01002"

	// Create a user
	_, err = client.CreateUser(&auth.CreateUserRequest{Name: userName})
	assert.Nil(t, err)
	// Add 2 miners
	success, err := client.UpsertMiner(userName, miner1)
	assert.Nil(t, err)
	assert.True(t, success)
	success, err = client.UpsertMiner(userName, miner2)
	assert.Nil(t, err)
	assert.True(t, success)

	return client, tmpDir
}

func testUpsertMiners(t *testing.T) {
	_, tmpDir := setupAndAddMiners(t)
	shutdown(t, tmpDir)
}

func testListMinerByUser(t *testing.T) {
	client, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	userName := "Rennbon"
	// List miner by user
	listResp, err := client.ListMiners(userName)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(listResp))
}

func testHasMiner(t *testing.T) {
	client, tmpDir := setupAndAddMiners(t)
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
}

func testMinerExistInUser(t *testing.T) {
	client, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	userName := "Rennbon"
	miner := "t01002"
	notExistMiner := "t010010"

	exist, err := client.MinerExistInUser(userName, miner)
	assert.Nil(t, err)
	assert.True(t, exist)

	exist, err = client.MinerExistInUser(userName, notExistMiner)
	assert.Nil(t, err)
	assert.False(t, exist)
}

func testGetUserByMiner(t *testing.T) {
	client, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	miner2 := "t01002"
	userName := "Rennbon"
	// Get user by miner
	getUserInfo, err := client.GetUserByMiner(&auth.GetUserByMinerRequest{Miner: miner2})
	assert.Nil(t, err)
	assert.Equal(t, userName, getUserInfo.Name)
}

func testDeleteMiner(t *testing.T) {
	client, tmpDir := setupAndAddMiners(t)
	defer shutdown(t, tmpDir)

	miner1 := "t01000"
	miner3 := "t01004"
	// Delete a miner
	success, err := client.DelMiner(miner1)
	assert.Nil(t, err)
	assert.True(t, success)

	// Check this miner
	has, err := client.HasMiner(&auth.HasMinerRequest{Miner: miner1})
	assert.Nil(t, err)
	assert.False(t, has)

	// Try to delete invalid miner
	success, err = client.DelMiner(miner3)
	assert.Nil(t, err)
	assert.False(t, success)
}
