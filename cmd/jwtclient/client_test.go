package jwtclient

import (
	"github.com/filecoin-project/venus-auth/auth"
	"os"
	"testing"
)

var MockCli *JWTClient

func TestMain(m *testing.M) {
	defer os.Exit(0)
	MockCli = NewJWTClient("http://localhost:8989")
	m.Run()
}

func TestClient_Verify(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	res, err := MockCli.Verify("miner-1", "mockSvc", "192.168.22.22", "192.168.22.21",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZH"+
			"lJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res)
}

func TestJWTClient_ListUsers(t *testing.T) {
	res, err := MockCli.ListUsers(auth.NewListUsersRequest(0, 20, 0, 0, 1+2))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}

//nolint
func BenchmarkJWTClient_Verify(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MockCli.Verify("miner-1", "mockSvc", "192.168.22.22", "192.168.22.21",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZH"+
				"lJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko")

	}
}
