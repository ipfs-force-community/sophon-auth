//stm: #integration
package integrate

import (
	"context"
	"testing"

	"github.com/filecoin-project/venus-auth/jwtclient"
	"github.com/stretchr/testify/assert"
)

func TestTokenApis(t *testing.T) {
	t.Run("generate token", testGenerateToken)
	t.Run("verify token", testVerifyToken)
	t.Run("list token", testListToken)
	t.Run("remove and recover token", testRemoveAndRecoverToken)
}

func setupAndGenerateToken(t *testing.T, name string, perm string) (*jwtclient.AuthClient, string, string) {
	server, tmpDir := setup(t)

	client, err := jwtclient.NewAuthClient(server.URL)
	assert.Nil(t, err)

	// Generate a token
	token, err := client.GenerateToken(name, perm, "")
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

	listResp, err := client.Tokens(int64(0), int64(10))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(listResp))
	assert.Equal(t, token, listResp[0].Token)
}

func testRemoveAndRecoverToken(t *testing.T) {
	name := "Rennbon"
	perm := "admin"

	client, tmpDir, token := setupAndGenerateToken(t, name, perm)
	defer shutdown(t, tmpDir)

	// Remove and then verify
	err := client.RemoveToken(token)
	assert.Nil(t, err)
	_, err = client.Verify(context.Background(), token)
	// Should not succeed this time
	assert.NotNil(t, err)

	// Recover this token and then verify
	err = client.RecoverToken(token)
	assert.Nil(t, err)
	verifyResp, err := client.Verify(context.Background(), token)
	// Should succeed this time
	assert.Nil(t, err)
	assert.Equal(t, name, verifyResp.Name)
	assert.Equal(t, perm, verifyResp.Perm)
}
