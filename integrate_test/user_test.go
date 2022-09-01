package integrate

import (
	"testing"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/stretchr/testify/assert"
)

func TestUserApis(t *testing.T) {
	t.Run("create user", testCreateUser)
	t.Run("get user", testGetUser)
	t.Run("has user", testHasUser)
	t.Run("list user", testListUser)
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
	_, tmpDir, _ := setupAndAddUser(t)
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
}

func testHasUser(t *testing.T) {
	client, tmpDir, createResp := setupAndAddUser(t)
	shutdown(t, tmpDir)

	// Has a user
	has, err := client.HasUser(&auth.HasUserRequest{Name: createResp.Name})
	assert.Nil(t, err)
	assert.True(t, has)

}

func testListUser(t *testing.T) {
	client, tmpDir, _ := setupAndAddUser(t)
	shutdown(t, tmpDir)

	// List users
	listResp, err := client.ListUsers(&auth.ListUsersRequest{
		Page: &core.Page{
			Skip:  0,
			Limit: 10,
		},
		State: int(core.UserStateUndefined),
	})
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

}
