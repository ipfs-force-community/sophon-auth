package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/ipfs-force-community/venus-auth/auth"
	"net/http"
)

type Client struct {
	cli *resty.Client
}

func NewClient(url string) *Client {
	client := resty.New().
		SetHostURL(url).
		SetHeader("Accept", "application/json")

	return &Client{
		cli: client,
	}
}

// Verify: post method for Verify token
// @spanId: local service unique Id
// @preHost: the IP of the request server
// @host: local service IP
// @token: jwt token gen from this service
func (c *Client) Verify(spanId, serviceName, preHost, host, token string) (*auth.VerifyResponse, error) {
	response, err := c.cli.SetHeader("X-Forwarded-For", host).
		SetHeader("X-Real-Ip", host).
		SetHeader("spanId", spanId).
		SetHeader("preHost", preHost).
		SetHeader("svcName", serviceName).
		SetHeader("Origin", host).
		R().
		SetFormData(map[string]string{
			"token": token,
		}).Post("/verify")
	if err != nil {
		return nil, err
	}
	switch response.StatusCode() {
	case http.StatusOK:
		return response.Result().(*auth.VerifyResponse), nil
	default:
		response.Result()
		return nil, fmt.Errorf("response code is : %d, msg:%s", response.StatusCode(), response.Body())
	}
}
