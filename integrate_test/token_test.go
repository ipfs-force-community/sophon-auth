package integrate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/jwtclient"
)

func TestTokenApis(t *testing.T) {
	t.Run("generate token", testGenerateToken)
	t.Run("verify token", testVerifyToken)
	t.Run("list token", testListToken)
	t.Run("remove and recover token", testRemoveAndRecoverToken)
}

func setupAndGenerateToken(t *testing.T, name string, perm string) (*jwtclient.AuthClient, string, string) {
	server, tmpDir, adminToken := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL, adminToken)
	assert.Nil(t, err)

	// Generate a token
	_, err = client.CreateUser(context.TODO(), &auth.CreateUserRequest{Name: name})
	assert.Nil(t, err)
	token, err := client.GenerateToken(context.TODO(), name, perm, "")
	assert.Nil(t, err)
	return client, tmpDir, token
}

func testGenerateToken(t *testing.T) {
	name := "Rennbon"
	perm := "admin"

	_, tmpDir, _ := setupAndGenerateToken(t, name, perm)
	shutdown(t, tmpDir)
}

func testVerifyToken(t *testing.T) {
	name := "Rennbon"
	perm := "admin"

	client, tmpDir, token := setupAndGenerateToken(t, name, perm)
	defer shutdown(t, tmpDir)

	verifyResp, err := client.Verify(context.Background(), token)
	assert.Nil(t, err)
	assert.Equal(t, name, verifyResp.Name)
	assert.Equal(t, perm, verifyResp.Perm)
}

func testListToken(t *testing.T) {
	name := "Rennbon"
	perm := "admin"

	client, tmpDir, token := setupAndGenerateToken(t, name, perm)
	defer shutdown(t, tmpDir)

	listResp, err := client.Tokens(context.TODO(), int64(0), int64(10))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(listResp))
	assert.Contains(t, []string{listResp[0].Token, listResp[1].Token}, token)
}

func testRemoveAndRecoverToken(t *testing.T) {
	name := "Rennbon"
	perm := "admin"

	client, tmpDir, token := setupAndGenerateToken(t, name, perm)
	defer shutdown(t, tmpDir)

	// Remove and then verify
	err := client.RemoveToken(context.TODO(), token)
	assert.Nil(t, err)
	_, err = client.Verify(context.Background(), token)
	// Should not succeed this time
	assert.NotNil(t, err)

	// Recover this token and then verify
	err = client.RecoverToken(context.TODO(), token)
	assert.Nil(t, err)
	verifyResp, err := client.Verify(context.Background(), token)
	// Should succeed this time
	assert.Nil(t, err)
	assert.Equal(t, name, verifyResp.Name)
	assert.Equal(t, perm, verifyResp.Perm)
}
