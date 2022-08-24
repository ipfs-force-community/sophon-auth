package integrate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
)

func TestSignerApis(t *testing.T) {
	t.Run("upsert signer", testUpsertSigners)
	t.Run("list miner by signer", testListSignerByUser)
	t.Run("has signer", testHasSigner)
	t.Run("get user by signer", testGetUserBySigner)
	t.Run("delete signer", testDeleteSigner)
}

func setupAndAddSigners(t *testing.T) (*jwtclient.AuthClient, string) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	userName := "test_user"
	userSigners := map[string][]string{
		"test_user": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
	}

	_, err = client.CreateUser(&auth.CreateUserRequest{Name: userName})
	assert.Nil(t, err)
	success, err := client.UpsertSigner(userName, userSigners[userName][0])
	assert.Nil(t, err)
	assert.True(t, success)
	success, err = client.UpsertSigner(userName, userSigners[userName][1])
	assert.Nil(t, err)
	assert.True(t, success)

	return client, tmpDir
}

func testUpsertSigners(t *testing.T) {
	_, tmpDir := setupAndAddSigners(t)
	shutdown(t, tmpDir)
}

func testListSignerByUser(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	userName := "test_user"
	// List miner by user
	listResp, err := client.ListSigners(userName)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(listResp))
}

func testHasSigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	userName := "test_user"
	userSigners := map[string][]string{
		"test_user": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
	}

	has, err := client.HasSigner(&auth.HasSignerRequest{Signer: userSigners[userName][0], User: userName})
	assert.Nil(t, err)
	assert.True(t, has)

	has, err = client.HasSigner(&auth.HasSignerRequest{Signer: userSigners[userName][1], User: ""})
	assert.Nil(t, err)
	assert.True(t, has)

	has, err = client.HasSigner(&auth.HasSignerRequest{Signer: "t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha", User: ""})
	assert.Nil(t, err)
	assert.False(t, has)
}

func testGetUserBySigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	userName := "test_user"
	userSigners := map[string][]string{
		"test_user": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
	}

	getUserInfo, err := client.GetUserBySigner(&auth.GetUserBySignerRequest{Signer: userSigners[userName][0]})
	assert.Nil(t, err)
	assert.Equal(t, userName, getUserInfo.Name)
}

func testDeleteSigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	userName := "test_user"
	userSigners := map[string][]string{
		"test_user": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
	}
	success, err := client.DelSigner(userSigners[userName][0])
	assert.Nil(t, err)
	assert.True(t, success)

	// Check this miner
	has, err := client.HasSigner(&auth.HasSignerRequest{Signer: userSigners[userName][0], User: ""})
	assert.Nil(t, err)
	assert.False(t, has)

	// Try to delete invalid miner
	success, err = client.DelSigner("f0128788")
	assert.Nil(t, err)
	assert.False(t, success)
}
