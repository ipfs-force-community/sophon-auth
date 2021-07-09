package cli

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"gotest.tools/assert"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/util"
)

var mockCnf *config.Config

//nolint
func TestMain(m *testing.M) {
	address.CurrentNetwork = address.Mainnet
	gin.SetMode(gin.TestMode)
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
	app, err := auth.NewOAuthApp(cnf.Secret, tmpPath, cnf.DB)
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
	go server.ListenAndServe() //nolint
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

func TestAccountBusiness(t *testing.T) {
	cli := mockClient(t)
	res1, err := cli.CreateAccount(&auth.CreateAccountRequest{
		Name:       "name1",
		Miner:      "f01234",
		Comment:    "this is a comment",
		State:      1,
		SourceType: 1,
	})
	if err != nil {
		t.Fatalf("create account err:%s", err)
	}
	t.Logf("account name: %s", res1.Name)

	res2, err := cli.CreateAccount(&auth.CreateAccountRequest{
		Name:       "name2",
		Miner:      "f02345",
		Comment:    "this is a comment",
		State:      1,
		SourceType: 1,
	})
	if err != nil {
		t.Fatalf("create account err:%s", err)
	}
	accounts, err := cli.ListAccounts(&auth.ListAccountsRequest{
		Page: &core.Page{
			Limit: 10,
			Skip:  0,
		},

		SourceType: 1,
		State:      0,
	})
	if err != nil {
		t.Fatalf("get tokens err:%s", err)
	}
	assert.DeepEqual(t, res1.Name, accounts[0].Name)
	assert.DeepEqual(t, res1.Miner.String(), accounts[0].Miner.String())
	assert.DeepEqual(t, res1.Comment, accounts[0].Comment)
	assert.DeepEqual(t, res1.State, accounts[0].State)
	assert.DeepEqual(t, res1.SourceType, accounts[0].SourceType)

	assert.DeepEqual(t, res2.Name, accounts[1].Name)
	assert.DeepEqual(t, res2.Miner.String(), accounts[1].Miner.String())
	assert.DeepEqual(t, res2.Comment, accounts[1].Comment)
	assert.DeepEqual(t, res2.State, accounts[1].State)
	assert.DeepEqual(t, res2.SourceType, accounts[1].SourceType)

	err = cli.UpdateAccount(&auth.UpdateAccountRequest{
		Name:       res1.Name,
		Miner:      "f01111",
		Comment:    "this is a comment?",
		State:      1,
		SourceType: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	miner, err := cli.GetMiner(&auth.GetMinerRequest{
		Miner: "f02345",
	})
	if err != nil {
		t.Fatalf("get miner err:%s", err)
	}
	assert.DeepEqual(t, accounts[1].Miner.String(), miner.Miner.String())

	has, err := cli.HasMiner(&auth.HasMinerRequest{
		Miner: "f02345",
	})
	if err != nil {
		t.Fatalf("has miner err:%s", err)
	}
	assert.DeepEqual(t, true, has)

	has, err = cli.HasMiner(&auth.HasMinerRequest{
		Miner: "f023452",
	})
	if err != nil {
		t.Fatalf("has miner err:%s", err)
	}
	assert.DeepEqual(t, false, has)

	account, err := cli.GetAccount(&auth.GetAccountRequest{
		Name: "name2",
	})
	if err != nil {
		t.Fatalf("get account err:%s", err)
	}
	assert.DeepEqual(t, accounts[1].Name, account.Name)
	assert.DeepEqual(t, accounts[1].Miner.String(), account.Miner.String())
}
