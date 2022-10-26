package storage

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var theStore Store
var cfg config.DBConfig

// badgerstore: go test -v ./storage/ -test.run TestStore --args -db=badger
// mysqlstore : go test -v ./storage/ -test.run TestStore --args -db=mysql -dns='root:ko2005@tcp(127.0.0.1:3306)/venus_auth?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s'
func TestMain(m *testing.M) {
	flag.StringVar(&cfg.Type, "db", "badger", "mysql or badger")
	flag.StringVar(&cfg.DSN, "dns", "", "sql connection string or badger data path")

	flag.Parse()

	if len(cfg.Type) == 0 {
		return
	}

	if err := setup(&cfg); err != nil {
		fmt.Printf("setup store(%s:%s) failed:%s\n", cfg.Type, cfg.DSN, err.Error())
		os.Exit(-1)
	}

	code := m.Run()
	if err := shutdown(); err != nil {
		fmt.Printf("shutdown failed:%s\n", err.Error())
	}
	os.Exit(code)
}

var userMiners = map[string]map[string]interface{}{
	"test_user_001": {"t01000": nil, "t01002": nil, "t01003": nil},
	"test_user_002": {"t01004": nil, "t01005": nil, "t01006": nil},
	"test_user_003": {"t01007": nil, "t01008": nil, "t01009": nil},
}

var limitStrs = `[{"Id":"794fc9a4-2b80-4503-835a-7e8e27360b3d","Name":"test_user_01","Service":"","API":"","ReqLimit":{"Cap":10,"ResetDur":120000000000}},{"Id":"252f581e-cbd2-4a61-a517-0b7df65013aa","Name":"test_user_02","Service":"","API":"","ReqLimit":{"Cap":10,"ResetDur":72000000000000}}]`
var originLimits []*UserRateLimit

var tokenStrs = `{"test-token-01":{"Name":"test-token-01","Perm":"admin","Secret":"d6234bf3f14a568a9c8315a6ee4f474e380beb2b65a64e6ba0142df72b454f4e","Extra":"","Token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwtdG9rZW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.DQ-ETWoEnNrpGKCikwZax6YUzdQIkhT0pHOTSta8770","CreateTime":"2022-03-18T16:11:53+08:00"}, "test-token-02":{"Name":"test-token-02","Perm":"admin","Secret":"862ed997d2943b7cd0997917f2ad524f4d56a4b50ff27e8bc680f4cc113cdd1b","Extra":"","Token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwtdG9rZW4tMDEiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.iw0L1UidBj0qaEddqc83AF36oa1lVeE9A_F9hXTK47c","CreateTime":"2022-03-18T16:14:49+08:00"}}`
var originTokens map[string]*KeyPair

func init() {
	if err := json.Unmarshal([]byte(tokenStrs), &originTokens); err != nil {
		panic(fmt.Sprintf("initialize origin Ratelimit failed:%s", err.Error()))
	}
	if err := json.Unmarshal([]byte(limitStrs), &originLimits); err != nil {
		panic(fmt.Sprintf("initialize origin Ratelimit failed:%s", err.Error()))
	}
}

func testAddUser(t *testing.T) {
	now := time.Now()
	for user := range userMiners {
		err := theStore.PutUser(&User{
			Id:         uuid.NewString(),
			Name:       user,
			UpdateTime: now,
			CreateTime: now,
		})
		if err != nil {
			if me, isok := err.(*mysql.MySQLError); isok {
				if me.Number == 1062 { // duplicate entry error is ok
					continue
				}
			}
			t.Fatalf("add user failed:%s", err.Error())
		}
	}
}

func testAddMiner(t *testing.T) {
	for u, ms := range userMiners {
		for m := range ms {
			addr, _ := address.NewFromString(m)
			_, err := theStore.UpsertMiner(addr, u, true)
			require.NoError(t, err)
		}
	}

}

func testListMiners(t *testing.T) {
	addr, _ := address.NewFromString("f01222345678999")
	// should be a 'not found error'
	_, err := theStore.GetUserByMiner(addr)
	require.Error(t, err)
	// make sure all miners(we just inserted) exist.
	for u, miners := range userMiners {
		ms, err := theStore.ListMiners(u)
		require.NoError(t, err)
		var minerMap = make(map[string]*Miner, len(ms))

		for _, m := range ms {
			minerMap[m.Miner.Address().String()] = m
		}

		for m := range miners {
			tmpMiner, isok := minerMap[m]
			require.Equal(t, isok, true)
			tmpUser, err := theStore.GetUserByMiner(tmpMiner.Miner.Address())
			require.NoError(t, err)
			require.Equal(t, tmpUser.Name, u)
		}
	}
}

func testDeleteUser(t *testing.T) {
	userName := "test_user_001"
	miners := userMiners[userName]

	res, err := theStore.GetUser(userName)
	require.Nil(t, err)
	require.Nil(t, theStore.DeleteUser(userName))
	has, err := theStore.HasUser(userName)
	require.Nil(t, err)
	require.False(t, has)
	_, err = theStore.GetUser(userName)
	require.Error(t, err)

	userRecord, err := theStore.GetUserRecord(userName)
	require.Nil(t, err)
	require.Equal(t, core.Deleted, userRecord.IsDeleted)
	userRecord.IsDeleted = core.NotDelete
	require.Equal(t, res, userRecord)

	finalMiner := address.Address{}
	for miner := range miners {
		addr, err := address.NewFromString(miner)
		require.Nil(t, err)
		has, err = theStore.HasMiner(addr)
		t.Log(addr.String())
		require.Nil(t, err)
		require.False(t, has)

		finalMiner = addr
	}
	_, err = theStore.GetUserByMiner(finalMiner)
	require.Error(t, err)
	list, err := theStore.ListMiners(userName)
	require.Nil(t, err)
	require.Len(t, list, 0)
}

func testDelMiners(t *testing.T) {
	for userName, miners := range userMiners {
		// already delete user
		if userName == "test_user_001" {
			continue
		}
		for m := range miners {
			addr, _ := address.NewFromString(m)
			_, err := theStore.DelMiner(addr)
			require.NoError(t, err)
		}
	}
}

func testTokens(t *testing.T) {
	// check put
	// token name the same
	nameSameToken := *originTokens["test-token-02"]
	nameSameToken.Secret = "93975e8ba920c7408fcfee2948b0929a86ba687894478bfcf1efbfa2c0b699bf"
	nameSameToken.Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiaGoiLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.9ZZP-acJY-PwluljFFIZ2ohIXaGoj4QzFrWTThSkAJk"
	for _, token := range originTokens {
		require.NoError(t, theStore.Put(token))
	}
	require.NoError(t, theStore.Put(&nameSameToken))

	// check get
	for _, token := range originTokens {
		old, err := theStore.Get(token.Token)
		require.NoError(t, err)
		require.Equal(t, token, old)
	}
	res, err := theStore.Get(nameSameToken.Token)
	require.NoError(t, err)
	require.Equal(t, &nameSameToken, res)

	// check get buy name
	kps, err := theStore.ByName("test-token-02")
	require.Nil(t, err)
	require.Len(t, kps, 2)

	// check list token, has, del
	tokens, err := theStore.List(0, 0)
	require.NoError(t, err)

	for _, token := range tokens {
		otk, exist := originTokens[token.Name]
		if !exist {
			continue
		}
		if token.Secret == nameSameToken.Secret {
			require.Equal(t, &nameSameToken, token)
		} else {
			require.Equal(t, token, otk)
		}

		has, err := theStore.Has(token.Token)
		require.NoError(t, err)
		require.Equal(t, has, true)

		require.NoError(t, theStore.Delete(token.Token))
		has, err = theStore.Has(token.Token)
		require.NoError(t, err)
		require.Equal(t, has, false)

		tokenRecord, err := theStore.GetTokenRecord(token.Token)
		require.Nil(t, err)
		require.Equal(t, core.Deleted, tokenRecord.IsDeleted)
		require.Equal(t, otk.Name, tokenRecord.Name)
	}
}

func testRatelimit(t *testing.T) {
	for _, l := range originLimits {
		_, err := theStore.PutRateLimit(l)
		require.NoError(t, err)
		limit, err := theStore.GetRateLimits(l.Name, "")
		require.NoError(t, err)
		require.Equal(t, l, limit[0])
		require.NoError(t, theStore.DelRateLimit(l.Name, l.Id))
		limit, err = theStore.GetRateLimits(l.Name, "")
		require.NoError(t, err)
		require.Equal(t, 0, len(limit))
	}
}

func TestStore(t *testing.T) {
	t.Run("add  users", testAddUser)
	t.Run("add miners", testAddMiner)
	t.Run("get miners", testListMiners)
	t.Run("del user", testDeleteUser)
	t.Run("del miners", testDelMiners)
	t.Run("test token", testTokens)
	t.Run("test ratelimit", testRatelimit)
}

func setup(cfg *config.DBConfig) error {
	var err error
	if cfg.Type == "badger" {
		if cfg.DSN, err = ioutil.TempDir("", "auth-datastore"); err != nil {
			return err
		}
		fmt.Printf("tmp badger store : %s\n", cfg.DSN)
	}
	theStore, err = NewStore(cfg, cfg.DSN)
	return err
}

func shutdown() error {
	if mysqlStore, isok := theStore.(*mysqlStore); isok {
		sqldb, err := mysqlStore.db.DB()
		if err != nil {
			return err
		}
		return sqldb.Close()
	}
	if badgerStroe, isok := theStore.(*badgerStore); isok {
		_ = badgerStroe.db.Close()
		fmt.Printf("shutdown, remove dir:%s\n", cfg.DSN)
		return os.RemoveAll(cfg.DSN)
	}
	return nil
}
