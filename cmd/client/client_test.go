package client

import (
	"os"
	"testing"
)

var MockCli *Client

func TestMain(m *testing.M) {
	defer os.Exit(0)
	MockCli = NewClient("http://localhost:8989")
	m.Run()
}

func TestClient_Verify(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	res, err := MockCli.Verify("miner-1", "192.168.22.22", "192.168.22.21", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.q3xz5oucOoT3xwMTct8pWMBrvhi_gizOz6QBgK-nOwc")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}
