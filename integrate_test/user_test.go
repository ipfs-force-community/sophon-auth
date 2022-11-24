// stm: #integration
package integrate

import (
	"testing"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/stretchr/testify/assert"
)

func TestUserApis(t *testing.T) {
	// stm: @VENUSAUTH_APP_BAD_RESPONSE_001, @VENUSAUTH_APP_SUCCESS_RESPONSE_001
	// stm: @VENUSAUTH_APP_CREATE_USER_001, @VENUSAUTH_APP_CREATE_USER_002, @VENUSAUTH_APP_CREATE_USER_003
	t.Run("create user", testCreateUser)
	// stm: @VENUSAUTH_APP_UPDATE_USER_001, @VENUSAUTH_APP_UPDATE_USER_002, @VENUSAUTH_APP_UPDATE_USER_003
	t.Run("update user", testUpdateUser)
	// stm: @VENUSAUTH_APP_GET_USER_001, @VENUSAUTH_APP_GET_USER_003
	t.Run("get user", testGetUser)
	// stm: @VENUSAUTH_APP_HAS_USER_001, @VENUSAUTH_APP_HAS_USER_002
	t.Run("has user", testHasUser)
	// stm: @VENUSAUTH_APP_LIST_USERS_001
	t.Run("list user", testListUser)
	// stm: @VENUSAUTH_APP_DELETE_USER_001, @VENUSAUTH_APP_DELETE_USER_002, @VENUSAUTH_APP_DELETE_USER_003
	// stm: @VENUSAUTH_APP_RECOVER_USER_001, @VENUSAUTH_APP_RECOVER_USER_003
	t.Run("delete user", testDeleteUser)
}

func setupAndAddUser(t *testing.T) (*jwtclient.AuthClient, string, *auth.CreateUserResponse) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	userName := "Rennbon"
	// Create a user
	createResp, err := client.CreateUser(&auth.CreateUserRequest{Name: userName})
	assert.Nil(t, err)
	assert.Equal(t, userName, createResp.Name)

	return client, tmpDir, createResp
}

func testCreateUser(t *testing.T) {
	c, tmpDir, userResp := setupAndAddUser(t)

	// user already exist error, and `BadResponse`
	_, err := c.CreateUser(&auth.CreateUserRequest{Name: userResp.Name})
	assert.Error(t, err)

	// `ShouldBind` failed
	_, err = c.CreateUser(&auth.CreateUserRequest{})
	assert.Error(t, err)

	shutdown(t, tmpDir)
}

func testGetUser(t *testing.T) {
	client, tmpDir, createResp := setupAndAddUser(t)
	shutdown(t, tmpDir)

	// Get a user
	getResp, err := client.GetUser(&auth.GetUserRequest{Name: createResp.Name})
	assert.Nil(t, err)
	assert.Equal(t, createResp.Name, getResp.Name)
	assert.Equal(t, createResp.Id, getResp.Id)
	assert.Equal(t, createResp.CreateTime, getResp.CreateTime)

	_, err = client.GetUser(&auth.GetUserRequest{Name: "not-exist-user"})
	assert.Error(t, err)
}

func testUpdateUser(t *testing.T) {
	c, tmpDir, user := setupAndAddUser(t)

	comment := "updated user comment"

	updateReq := &auth.UpdateUserRequest{Name: user.Name, Comment: &comment, State: core.UserStateEnabled}
	err := c.UpdateUser(updateReq)
	assert.NoError(t, err)

	// `ShouldBind` failed
	err = c.UpdateUser(&auth.UpdateUserRequest{})
	assert.Error(t, err)

	// user not exist error
	err = c.UpdateUser(&auth.UpdateUserRequest{Name: "not-exist-user-name"})
	assert.Error(t, err)

	shutdown(t, tmpDir)
}

func testHasUser(t *testing.T) {
	client, tmpDir, createResp := setupAndAddUser(t)
	shutdown(t, tmpDir)

	// Has a user
	has, err := client.HasUser(&auth.HasUserRequest{Name: createResp.Name})
	assert.Nil(t, err)
	assert.True(t, has)
	// `ShouldBind` failed
	_, err = client.HasUser(&auth.HasUserRequest{})
	assert.Error(t, err)
}

func testListUser(t *testing.T) {
	client, tmpDir, _ := setupAndAddUser(t)
	shutdown(t, tmpDir)

	// List users
	listResp, err := client.ListUsers(auth.NewListUsersRequest(0, 10, int(core.UserStateUndefined)))
	assert.Nil(t, err)
	assert.Equal(t, len(listResp), 1)
}

func testDeleteUser(t *testing.T) {
	client, tmpDir, createResp := setupAndAddUser(t)
	shutdown(t, tmpDir)

	userName := createResp.Name

	// Delete user and then try to call get and has
	err := client.DeleteUser(&auth.DeleteUserRequest{Name: userName})
	assert.Nil(t, err)
	// Get should fail
	_, err = client.GetUser(&auth.GetUserRequest{Name: userName})
	assert.NotNil(t, err)
	// Has should return false
	has, err := client.HasUser(&auth.HasUserRequest{Name: userName})
	assert.Nil(t, err)
	assert.False(t, has)

	// Recover the user and check
	err = client.RecoverUser(&auth.RecoverUserRequest{Name: userName})
	assert.Nil(t, err)
	has, err = client.HasUser(&auth.HasUserRequest{Name: userName})
	assert.Nil(t, err)
	assert.True(t, has)

	// Recover not exist user.
	err = client.RecoverUser(&auth.RecoverUserRequest{Name: "not-exist-user"})
	assert.Error(t, err)

	// `ShouldBind` failed
	err = client.DeleteUser(&auth.DeleteUserRequest{})
	assert.Error(t, err)

	// Delete a not exists user
	err = client.DeleteUser(&auth.DeleteUserRequest{Name: "not-exist-user"})
	assert.Error(t, err)
}
