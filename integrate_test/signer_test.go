package integrate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-address"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
)

var userSignerAddrs = getUserSignerAddrs()

func getUserSignerAddrs() map[string][]address.Address {
	ret := make(map[string][]address.Address)

	user2signerCase := map[string][]string{
		"test_user01": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q"},
		"test_user02": {"t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua", "t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha"},
	}

	for user, signers := range user2signerCase {
		var sAddrs []address.Address
		for _, s := range signers {
			addr, err := address.NewFromString(s)
			if err != nil {
				panic(err)
			}
			sAddrs = append(sAddrs, addr)
		}
		ret[user] = sAddrs
	}
	return ret
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
	server, tmpDir, token := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL, token)
	assert.Nil(t, err)
	for username, signers := range userSignerAddrs {
		_, err = client.CreateUser(context.TODO(), &auth.CreateUserRequest{Name: username})
		assert.Nil(t, err)
		err = client.RegisterSigners(context.Background(), username, signers)
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
	for user, signers := range userSignerAddrs {
		for _, signer := range signers {
			bExist, err := client.SignerExistInUser(context.Background(), user, signer)
			assert.Nil(t, err)
			assert.True(t, bExist)
		}
	}
}

func testListSignerByUser(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	for user, signers := range userSignerAddrs {
		ss, err := client.ListSigners(context.Background(), user)
		assert.Nil(t, err)

		ns := make([]address.Address, len(ss))
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
	for _, signers := range userSignerAddrs {
		for _, signer := range signers {
			bExist, err := client.HasSigner(context.Background(), signer)
			assert.Nil(t, err)
			assert.True(t, bExist)
		}
	}
}

func testGetUserBySigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	signer, err := address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	assert.Nil(t, err)
	users, err := client.GetUserBySigner(context.Background(), signer)
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
	signer, err := address.NewFromString("t1sgeoaugenqnzftqp7wvwqebcozkxa5y7i56sy2q")
	assert.Nil(t, err)

	err = client.UnregisterSigners(context.Background(), userName, []address.Address{signer})
	assert.Nil(t, err)

	bExist, err := client.SignerExistInUser(context.Background(), userName, signer)
	assert.Nil(t, err)
	assert.False(t, bExist)
}

func testDeleteSigner(t *testing.T) {
	client, tmpDir := setupAndAddSigners(t)
	defer shutdown(t, tmpDir)

	signer := "t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua"
	signerAddr, err := address.NewFromString("t15rynkupqyfx5ebvaishg7duutwb5ooq2qpaikua")
	assert.Nil(t, err)

	bDel, err := client.DelSigner(context.TODO(), signer)
	assert.Nil(t, err)
	assert.True(t, bDel)

	has, err := client.HasSigner(context.Background(), signerAddr)
	assert.Nil(t, err)
	assert.False(t, has)

	// delete again
	bDel, err = client.DelSigner(context.TODO(), signer)
	assert.Nil(t, err)
	assert.False(t, bDel)
}
