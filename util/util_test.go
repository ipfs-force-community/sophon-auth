package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWTPayloadMap(t *testing.T) {
	// valid token
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.cThIIoDvwdueQB468K5xDc5633seEFoqwxjF_xSJyQQ"
	result, err := JWTPayloadMap(token)
	assert.Nil(t, err)
	assert.Equal(t, "John Doe", result["name"].(string))
	assert.Equal(t, "1234567890", result["sub"].(string))
	assert.Equal(t, 1516239022, int(result["iat"].(float64)))

	// invalid token
	token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaF0IjoxNTE2MjM5MDIyfQ.cThIIoDvwdueQB468K5xDc5633seEFoqwxjF_xSJyQQ"
	_, err = JWTPayloadMap(token)
	assert.NotNil(t, err)
}
