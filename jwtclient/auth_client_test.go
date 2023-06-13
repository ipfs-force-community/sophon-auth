// stm: #unit
package jwtclient

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/util"
)

var cli *AuthClient

// nolint
func TestMain(m *testing.M) {
	cnf := config.DefaultConfig()
	flag.StringVar(&cnf.DB.Type, "db", "badger", "mysql or badger")
	flag.StringVar(&cnf.DB.DSN, "dns", "", "sql connection string or badger data path")
	flag.Parse()

	address.CurrentNetwork = address.Mainnet
	gin.SetMode(gin.TestMode)
	port, err := util.GetAvailablePort()
	if err != nil {
		log.Fatalf("failed to get available port err:%s", err)
	}
	cnf.Listen = fmt.Sprintf("127.0.0.1:%d", port)

	var tmpPath string
	if cnf.DB.Type == "badger" {
		if tmpPath, err = os.MkdirTemp("", "auth-serve"); err != nil {
			log.Fatalf("failed to create temp dir err:%s", err)
		}
	}

	// stm: @VENUSAUTH_JWT_NEW_OAUTH_SERVICE_001
	app, err := auth.NewOAuthApp(tmpPath, cnf.DB)
	if err != nil {
		log.Fatalf("Failed to init oauthApp : %s", err)
	}
	token, err := app.GetDefaultAdminToken()
	if err != nil {
		log.Fatalf("Failed to get default admin token : %s", err)
	}

	router := auth.InitRouter(app)
	server := &http.Server{
		Addr:         cnf.Listen,
		Handler:      router,
		ReadTimeout:  cnf.ReadTimeout,
		WriteTimeout: cnf.WriteTimeout,
		IdleTimeout:  cnf.IdleTimeout,
	}

	go func() {
		log.Infof("server start and listen on %s", cnf.Listen)
		if err = server.ListenAndServe(); err != nil {
			log.Errorf("start serve failed: %v", err)
		}
	}() //nolint

	if cli, err = NewAuthClient("http://"+cnf.Listen, token); err != nil {
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
	ctx := context.TODO()
	var originTks []string
	_, err := cli.CreateUser(context.TODO(), &auth.CreateUserRequest{
		Name: "Rennbon1",
	})
	if err != nil {
		t.Fatalf("create user err:%s", err)
	}
	tk1, err := cli.GenerateToken(context.TODO(), "Rennbon1", core.PermAdmin, "custom params")
	if err != nil {
		t.Fatalf("gen token err:%s", err)
	}
	originTks = append(originTks, tk1)

	_, err = cli.CreateUser(context.TODO(), &auth.CreateUserRequest{
		Name: "Rennbon2",
	})
	if err != nil {
		t.Fatalf("create user err:%s", err)
	}
	tk2, err := cli.GenerateToken(context.TODO(), "Rennbon2", core.PermRead, "custom params")
	if err != nil {
		t.Fatalf("gen token err:%s", err)
	}
	originTks = append(originTks, tk2)

	tks, err := cli.Tokens(ctx, 0, 0)
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

	err = cli.RemoveToken(ctx, tk1)
	if err != nil {
		t.Fatalf("remove token err:%s", err)
	}
	tks2, err := cli.Tokens(ctx, 0, 0)
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
		resp, err := cli.CreateUser(context.TODO(), req)
		if err != nil {
			// user already exists error is ok
			if strings.Index(err.Error(), "already exists") > 0 {
				resp, err := cli.GetUser(context.Background(), req.Name)
				assert.NoError(t, err)
				originUsers[resp.Id] = resp
				continue
			}
			t.Fatalf("create user err:%s", err)
		}
		originUsers[resp.Id] = resp
	}

	users, err := cli.ListUsers(context.Background(), 0, 10, core.UserStateUndefined)
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
		assert.Equal(t, u.Name, tmpU.Name)
		assert.Equal(t, u.Comment, tmpU.Comment)
		assert.Equal(t, u.State, tmpU.State)
	}

	newComment := "this is a new comment"
	for _, res1 := range originUsers {
		err = cli.UpdateUser(context.TODO(), &auth.UpdateUserRequest{
			Name:    res1.Name,
			Comment: &newComment,
			State:   1,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, err = cli.UpsertMiner(context.TODO(), res1.Name, "f02345", true)
		assert.NoError(t, err)
		break
	}

	mAddr1, err := address.NewFromString("f02345")
	assert.NoError(t, err)
	mAddr2, err := address.NewFromString("f023452")
	assert.NoError(t, err)

	user, err := cli.GetUserByMiner(context.Background(), mAddr1)
	if err != nil {
		t.Fatalf("get miner err:%s", err)
	}

	has, err := cli.HasMiner(context.Background(), mAddr1)
	if err != nil {
		fmt.Printf("err: %s\n", err.Error())
	}
	assert.Equal(t, true, has)

	has, err = cli.HasMiner(context.Background(), mAddr2)
	if err != nil {
		t.Fatalf("has miner err:%s", err)
	}
	assert.Equal(t, false, has)

	exist, err := cli.MinerExistInUser(context.Background(), user.Name, mAddr1)
	if err != nil {
		t.Fatalf("check miner exist in user err:%s", err)
	}
	assert.Equal(t, true, exist)

	exist, err = cli.MinerExistInUser(context.Background(), user.Name, mAddr2)
	if err != nil {
		t.Fatalf("check miner exist in user err:%s", err)
	}
	assert.Equal(t, false, exist)

	user, err = cli.GetUser(context.Background(), "name2")
	if err != nil {
		t.Fatalf("get user err:%s", err)
	}
	assert.Equal(t, "name2", user.Name)

	err = cli.VerifyUsers(context.Background(), []string{"name1", "name2"})
	if err != nil {
		t.Fatalf("verify users err:%s", err)
	}
}

func TestClient_Verify(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}

	kps, err := cli.Tokens(context.TODO(), 0, 10)
	if err != nil {
		t.Fatalf("get key-pars failed:%s", err.Error())
	}

	ctx := context.TODO()
	for _, kp := range kps {
		res, err := cli.Verify(ctx, kp.Token)
		assert.NoError(t, err)
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
	res, err := cli.ListUsers(context.Background(), 0, 20, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res {
		t.Log(v)
	}
}
