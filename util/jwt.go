package util

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func DecodeToBytes(enc []byte) ([]byte, error) {
	encoding := base64.RawURLEncoding
	dec := make([]byte, encoding.DecodedLen(len(enc)))
	if _, err := encoding.Decode(dec, enc); err != nil {
		return nil, err
	}
	return dec, nil
}

func JWTPayloadMap(token string) (map[string]interface{}, error) {
	tokenSpan := strings.Split(token, ".")
	payload := []byte(tokenSpan[1])
	pb, err := DecodeToBytes(payload)
	if err != nil {
		return nil, err
	}
	pMap := map[string]interface{}{}
	err = json.Unmarshal(pb, &pMap)
	if err != nil {
		return nil, err
	}
	return pMap, nil
}
