package util

import (
	"encoding/base64"
	"encoding/json"
	"net"
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

func MacAddr() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("net interfaces" + err.Error())
	}
	mac := ""
	for _, netInterface := range interfaces {
		mac = netInterface.HardwareAddr.String()
		if len(mac) == 0 {
			continue
		}
		break
	}
	return mac
}
