package integrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/jwtclient"
)

var userSigners = map[string][]string{
	"test_user01": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
	"test_user02": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha"},
}

func TestSignerAPI(t *testing.T) {
	t.Run("register signer", testRegisterSigners)
	t.Run("signer exist in user", testSignerExistInUser)
	t.Run("list signer of user", testListSignerByUser)
	t.Run("has signer", testHasSigner)
	t.Run("get user by signer", testGetUserBySigner)
	t.Run("unregister signer", testUnregisterSigner)
	t.Run("delete signer", testDeleteSigner)
}

func setupAndAddSigners(t *testing.T) (*jwtclient.AuthClient, string) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	for username, signers := range userSigners {
		_, err = client.CreateUser(&auth.CreateUserRequest{Name: username})
		assert.Nil(t, err)

		err = client.RegisterSigners(username, signers)
		assert.Nil(t, err)
	}

	return client, tmpDir
}

func testRegisterSigners(t *testing.T) {
	_, tmpDir := setupAndAddSigners(t)
	shutdown(t, tmpDir)
}

func testSignerExistInUser(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	for user, signers := range userSigners {
		for _, signer := range signers {
			bExist, err := client.SignerExistInUser(user, signer)
			assert.Nil(t, err)
			assert.True(t, bExist)
		}
	}
}

func testListSignerByUser(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	for user, signers := range userSigners {
		ss, err := client.ListSigners(user)
		assert.Nil(t, err)

		ns := make([]string, len(ss))
		for idx, s := range ss {
			ns[idx] = s.Signer
		}

		for _, signer := range signers {
			require.Contains(t, ns, signer)
		}
	}
}

func testHasSigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	for _, signers := range userSigners {
		for _, signer := range signers {
			bExist, err := client.HasSigner(signer)
			assert.Nil(t, err)
			assert.True(t, bExist)
		}
	}
}

func testGetUserBySigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	signer := "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"
	users, err := client.GetUserBySigner(signer)
	assert.Nil(t, err)

	names := make([]string, len(users))
	for idx, user := range users {
		names[idx] = user.Name
	}
	require.Contains(t, names, "test_user01")
	require.Contains(t, names, "test_user02")
}

func testUnregisterSigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	userName := "test_user01"
	signer := "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"

	err := client.UnregisterSigners(userName, []string{signer})
	assert.Nil(t, err)

	bExist, err := client.SignerExistInUser(userName, signer)
	assert.Nil(t, err)
	assert.False(t, bExist)
}

func testDeleteSigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	signer := "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"

	bDel, err := client.DelSigner(signer)
	assert.Nil(t, err)
	assert.True(t, bDel)

	has, err := client.HasSigner(signer)
	assert.Nil(t, err)
	assert.False(t, has)

	// delete again
	bDel, err = client.DelSigner(signer)
	assert.Nil(t, err)
	assert.False(t, bDel)
}
