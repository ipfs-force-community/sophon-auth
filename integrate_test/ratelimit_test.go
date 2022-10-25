// stm: #integration
package integrate

import (
	"testing"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitApis(t *testing.T) {
	// stm: @VENUSAUTH_APP_ADD_USER_RATE_LIMIT_001, @VENUSAUTH_APP_ADD_USER_RATE_LIMIT_002, @VENUSAUTH_APP_ADD_USER_RATE_LIMIT_003
	// stm: @VENUSAUTH_APP_UPSERT_USER_RATE_LIMIT_001, @VENUSAUTH_APP_UPSERT_USER_RATE_LIMIT_002
	t.Run("upsert rate limit", testUpsertUserRateLimit)
	// stm: @VENUSAUTH_APP_GET_USER_RATE_LIMIT_001, @VENUSAUTH_APP_GET_USER_RATE_LIMIT_002
	t.Run("get rate limit", testGetRateLimit)
	// stm: @VENUSAUTH_APP_DEL_USER_RATE_LIMIT_001, @VENUSAUTH_APP_DEL_USER_RATE_LIMIT_003
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
	c, tmpDir := setupAndAddRateLimits(t)

	// `ShouldBind` failed
	_, err := c.UpsertUserRateLimit(&auth.UpsertUserRateLimitReq{})
	assert.Error(t, err)

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

	// `ShouldBind` failed
	_, err = client.GetUserRateLimit("", "")
	assert.Error(t, err)
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

	// if there is an error deleting user rate limits
	_, err = client.DelUserRateLimit(&auth.DelUserRateLimitReq{})
	assert.Error(t, err)
}
