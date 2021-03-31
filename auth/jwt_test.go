package auth

import (
	"encoding/json"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestTokenDecode(t *testing.T) {
	payload := []byte("eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ")
	pb, err := DecodeToBytes(payload)
	if err != nil {
		t.Fatal(err)
	}
	a := map[string]interface{}{}
	err = json.Unmarshal(pb, &a)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, a["name"], "John Doe")
}
