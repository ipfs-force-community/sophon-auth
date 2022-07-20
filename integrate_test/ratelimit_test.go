package integrate

import (
	"testing"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/cmd/jwtclient"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitApis(t *testing.T) {
	t.Run("upsert rate limit", testUpsertUserRateLimit)
	t.Run("get rate limit", testGetRateLimit)
	t.Run("delete rate limit", testDeleteRateLimit)
}

func setupAndAddRateLimits(t *testing.T) (*jwtclient.AuthClient, string) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	// Create a user
	userName := "Rennbon"
	_, err = client.CreateUser(&auth.CreateUserRequest{Name: userName})
	assert.Nil(t, err)

	// Insert rate limit
	upsertReq := auth.UpsertUserRateLimitReq{
		Id:      "794fc9a4-2b80-4503-835a-7e8e27360b3d",
		Name:    userName,
		Service: "",
		API:     "",
		ReqLimit: storage.ReqLimit{
			Cap:      10,
			ResetDur: 120000000000,
		},
	}

	upsertResp, err := client.UpsertUserRateLimit(&upsertReq)
	assert.Nil(t, err)
	assert.Equal(t, upsertReq.Id, upsertResp)

	return client, tmpDir
}

func testUpsertUserRateLimit(t *testing.T) {
	_, tmpDir := setupAndAddRateLimits(t)
	shutdown(t, tmpDir)
}

func testGetRateLimit(t *testing.T) {
	client, tmpDir := setupAndAddRateLimits(t)
	defer shutdown(t, tmpDir)

	userName := "Rennbon"
	reqId := "794fc9a4-2b80-4503-835a-7e8e27360b3d"
	// Get user rate limit
	getResp, err := client.GetUserRateLimit(userName, reqId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(getResp))
	assert.Equal(t, reqId, getResp[0].Id)
}

func testDeleteRateLimit(t *testing.T) {
	client, tmpDir := setupAndAddRateLimits(t)
	defer shutdown(t, tmpDir)

	userName := "Rennbon"
	reqId := "794fc9a4-2b80-4503-835a-7e8e27360b3d"
	// Delete rate limit
	deleteResp, err := client.DelUserRateLimit(&auth.DelUserRateLimitReq{Name: userName, Id: reqId})
	assert.Nil(t, err)
	assert.Equal(t, deleteResp, reqId)

	// Try to get deleted rate limit
	getResp, err := client.GetUserRateLimit(userName, reqId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(getResp))
}
