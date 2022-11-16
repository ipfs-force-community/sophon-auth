// stm: #unit
package jwtclient

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
	"gotest.tools/assert"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/util"
)

var cli *AuthClient

// nolint
func TestMain(m *testing.M) {
	cnf, err := config.DefaultConfig()
	if err != nil {
		log.Fatalf("failed to get default config err:%s", err)
	}
	flag.StringVar(&cnf.DB.Type, "db", "badger", "mysql or badger")
	flag.StringVar(&cnf.DB.DSN, "dns", "", "sql connection string or badger data path")
	flag.Parse()

	address.CurrentNetwork = address.Mainnet
	gin.SetMode(gin.TestMode)
	port, err := util.GetAvailablePort()
	if err != nil {
		log.Fatalf("failed to get available port err:%s", err)
	}
	cnf.Port = strconv.FormatInt(int64(port), 10)
	var tmpPath string

	if cnf.DB.Type == "badger" {
		if tmpPath, err = ioutil.TempDir("", "auth-serve"); err != nil {
			log.Fatalf("failed to create temp dir err:%s", err)
		}
		defer os.RemoveAll(tmpPath)
	}

	// stm: @VENUSAUTH_JWT_NEW_OAUTH_SERVICE_001
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

	go func() {
		log.Infof("server start and listen on %s", cnf.Port)
		_ = server.ListenAndServe()
	}() //nolint

	if cli, err = NewAuthClient("http://localhost:" + cnf.Port); err != nil {
		log.Fatalf("create auth client failed:%s\n", err.Error())
		return
	}

	defer func() {
		log.Info("close auth service")
		_ = server.Close()
	}()

	m.Run()
}

func TestTokenBusiness(t *testing.T) {
	var originTks []string
	tk1, err := cli.GenerateToken("Rennbon1", core.PermAdmin, "custom params")
	if err != nil {
		t.Fatalf("gen token err:%s", err)
	}
	originTks = append(originTks, tk1)

	tk2, err := cli.GenerateToken("Rennbon2", core.PermRead, "custom params")
	if err != nil {
		t.Fatalf("gen token err:%s", err)
	}
	originTks = append(originTks, tk2)

	tks, err := cli.Tokens(0, 0)
	if err != nil {
		t.Fatalf("get tokens err:%s", err)
	}

	listTks := make(map[string]*auth.TokenInfo)

	for _, tkInfo := range tks {
		listTks[tkInfo.Token] = tkInfo
	}

	for _, tk := range originTks {
		_, find := listTks[tk]
		assert.Equal(t, find, true)
	}

	err = cli.RemoveToken(tk1)
	if err != nil {
		t.Fatalf("remove token err:%s", err)
	}
	tks2, err := cli.Tokens(0, 0)
	if err != nil {
		t.Fatalf("get tokens err:%s", err)
	}

	listTks = make(map[string]*auth.TokenInfo)
	for _, tkInfo := range tks2 {
		listTks[tkInfo.Token] = tkInfo
	}

	// tk1 already been deleted, only tk2 is left
	_, find := listTks[tk1]
	assert.Equal(t, find, false)
	_, find = listTks[tk2]
	assert.Equal(t, find, true)
}

func TestUserBusiness(t *testing.T) {
	comment := "this is a comment"
	createReqs := []*auth.CreateUserRequest{
		{
			Name:    "name1",
			Comment: nil,
			State:   1,
		},
		{
			Name:    "name2",
			Comment: &comment,
			State:   1,
		},
	}
	originUsers := make(map[string]*auth.CreateUserResponse, len(createReqs))
	var err error
	for _, req := range createReqs {
		resp, err := cli.CreateUser(req)
		if err != nil {
			// user already exists error is ok
			if strings.Index(err.Error(), "already exists") > 0 {
				resp, err := cli.GetUser(&auth.GetUserRequest{Name: req.Name})
				assert.NilError(t, err)
				originUsers[resp.Id] = resp
				continue
			}
			t.Fatalf("create user err:%s", err)
		}
		originUsers[resp.Id] = resp
	}

	users, err := cli.ListUsers(&auth.ListUsersRequest{
		Page: &core.Page{
			Limit: 10,
		},
		State: int(core.UserStateUndefined),
	})
	if err != nil {
		t.Fatalf("get tokens err:%s", err)
	}

	listUserMaps := make(map[string]*auth.OutputUser, len(users))
	for _, u := range users {
		listUserMaps[u.Id] = u
	}

	for id, u := range originUsers {
		tmpU, find := listUserMaps[id]
		assert.Equal(t, find, true)
		assert.DeepEqual(t, u.Name, tmpU.Name)
		assert.DeepEqual(t, u.Comment, tmpU.Comment)
		assert.DeepEqual(t, u.State, tmpU.State)
	}

	newComment := "this is a new comment"
	for _, res1 := range originUsers {
		err = cli.UpdateUser(&auth.UpdateUserRequest{
			Name:    res1.Name,
			Comment: &newComment,
			State:   1,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = cli.UpsertMiner(res1.Name, "f02345", true)
		assert.NilError(t, err)
		break
	}

	user, err := cli.GetUserByMiner(&auth.GetUserByMinerRequest{
		Miner: "f02345",
	})
	if err != nil {
		t.Fatalf("get miner err:%s", err)
	}

	has, err := cli.HasMiner(&auth.HasMinerRequest{
		Miner: "f02345",
	})
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
	}
	assert.DeepEqual(t, true, has)

	has, err = cli.HasMiner(&auth.HasMinerRequest{
		Miner: "f023452",
	})
	if err != nil {
		t.Fatalf("has miner err:%s", err)
	}
	assert.DeepEqual(t, false, has)

	exist, err := cli.MinerExistInUser(user.Name, "f02345")
	if err != nil {
		t.Fatalf("check miner exist in user err:%s", err)
	}
	assert.DeepEqual(t, true, exist)

	exist, err = cli.MinerExistInUser(user.Name, "f023452")
	if err != nil {
		t.Fatalf("check miner exist in user err:%s", err)
	}
	assert.DeepEqual(t, false, exist)

	user, err = cli.GetUser(&auth.GetUserRequest{
		Name: "name2",
	})
	if err != nil {
		t.Fatalf("get user err:%s", err)
	}
	assert.DeepEqual(t, users[1].Name, user.Name)

	err = cli.VerifyUsers([]string{"name1", "name2"})
	if err != nil {
		t.Fatalf("verify users err:%s", err)
	}
}

func TestClient_Verify(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}

	kps, err := cli.Tokens(0, 10)
	if err != nil {
		t.Fatalf("get key-pars failed:%s", err.Error())
	}

	ctx := context.TODO()
	for _, kp := range kps {
		res, err := cli.Verify(ctx, kp.Token)
		assert.NilError(t, err)
		assert.Equal(t, res.Name, kp.Name)
		assert.Equal(t, res.Perm, kp.Perm)
	}

	if _, err = cli.Verify(context.TODO(),
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiUmVubmJvbiIsInBlcm0iOiJhZG1pbiIsImV4dCI6ImV5SkJiR3h2ZH"+
			"lJNld5SnlaV0ZrSWl3aWQzSnBkR1VpTENKemFXZHVJaXdpWVdSdGFXNGlYWDAifQ.gONkC1v8AuY-ZP2WhU62EonWmyPeOW1pFhnRM-Fl7ko"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expect error with 'not found' message")
	}
}

func TestJWTClient_ListUsers(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	res, err := cli.ListUsers(auth.NewListUsersRequest(0, 20, 1))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}
