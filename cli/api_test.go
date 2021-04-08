package cli

import (
	"github.com/ipfs-force-community/venus-auth/auth"
	"github.com/ipfs-force-community/venus-auth/config"
	"github.com/ipfs-force-community/venus-auth/core"
	"github.com/ipfs-force-community/venus-auth/util"
	"gotest.tools/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
)

var mockCnf *config.Config

func TestMain(m *testing.M) {
	cnf, err := config.DefaultConfig()
	if err != nil {
		log.Fatalf("failed to get default config err:%s", err)
	}
	port, err := util.GetAvailablePort()
	if err != nil {
		log.Fatalf("failed to get available port err:%s", err)
	}
	cnf.Port = strconv.FormatInt(int64(port), 10)
	mockCnf = cnf
	tmpPath, err := ioutil.TempDir("", "auth-serve")
	if err != nil {
		log.Fatalf("failed to create temp dir err:%s", err)
	}
	defer os.RemoveAll(tmpPath)
	app, err := auth.NewOAuthApp(cnf.Secret, tmpPath)
	if err != nil {
		log.Fatalf("Failed to init oauthApp : %s", err)
	}
	router := auth.InitRouter(app)
	server := &http.Server{
		Addr:         ":" + cnf.Port,
		Handler:      router,
		ReadTimeout:  cnf.ReadTimeout,
		WriteTimeout: cnf.WriteTimeout,
		IdleTimeout:  cnf.IdleTimeout,
	}
	log.Printf("server start and listen on %s", cnf.Port)
	go server.ListenAndServe()
	defer func() {
		server.Close()
		log.Println("server closed")
	}()
	m.Run()
}

func mockClient(t *testing.T) *localClient {
	cli, err := newClient(mockCnf.Port)
	if err != nil {
		t.Fatal(err)
	}
	return cli
}

func TestTokenBusiness(t *testing.T) {
	cli := mockClient(t)
	tk1, err := cli.GenerateToken("Rennbon1", core.PermAdmin, "custom params")
	if err != nil {
		t.Fatalf("gen token err:%s", err)
	}
	t.Logf("gen token: %s", tk1)

	tk2, err := cli.GenerateToken("Rennbon2", core.PermRead, "custom params")
	if err != nil {
		t.Fatalf("gen token err:%s", err)
	}
	tks, err := cli.Tokens(0, 10)
	if err != nil {
		t.Fatalf("get tokens err:%s", err)
	}
	assert.DeepEqual(t, tk1, tks[0].Token)
	assert.DeepEqual(t, tk2, tks[1].Token)

	err = cli.RemoveToken(tk1)
	if err != nil {
		t.Fatalf("remove token err:%s", err)
	}
	tks2, err := cli.Tokens(0, 10)
	if err != nil {
		t.Fatalf("get tokens err:%s", err)
	}
	assert.Equal(t, len(tks2), 1)
	assert.DeepEqual(t, tks2[0].Token, tk2)
}
