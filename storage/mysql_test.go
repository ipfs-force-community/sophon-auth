package storage

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/filecoin-project/go-address"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/filecoin-project/venus-auth/core"
)

type anyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func wrapper(f func(*testing.T, *mysqlStore, sqlmock.Sqlmock), mySQLStore *mysqlStore, mock sqlmock.Sqlmock) func(t *testing.T) {
	return func(t *testing.T) {
		f(t, mySQLStore, mock)
	}
}

func TestMysqlStore(t *testing.T) {
	mySQLStore, mock, sqlDB, err := mysqlSetup()
	if err != nil {
		t.Fatal(err)
	}

	// Token
	t.Run("mysql put token", wrapper(testMySQLPutToken, mySQLStore, mock))
	t.Run("mysql update token", wrapper(testMySQLUpdateToken, mySQLStore, mock))
	t.Run("mysql has token", wrapper(testMySQLHasToken, mySQLStore, mock))
	t.Run("mysql get token", wrapper(testMySQLGetToken, mySQLStore, mock))
	t.Run("mysql list tokens", wrapper(testMySQLListTokens, mySQLStore, mock))
	t.Run("mysql get token record", wrapper(testMySQLGetTokenRecord, mySQLStore, mock))
	t.Run("mysql get token by name", wrapper(testMySQLGetTokenByName, mySQLStore, mock))
	t.Run("mysql delete token", wrapper(testMySQLDeleteToken, mySQLStore, mock))

	// User
	t.Run("mysql put users", wrapper(testMySQLPutUser, mySQLStore, mock))
	t.Run("mysql update user", wrapper(testMySQLUpdateUser, mySQLStore, mock))
	t.Run("mysql has user", wrapper(testMySQLHasUser, mySQLStore, mock))
	t.Run("mysql get user", wrapper(testMySQLGetUser, mySQLStore, mock))
	t.Run("mysql get user record", wrapper(testMySQLGetUserRecord, mySQLStore, mock))
	t.Run("mysql list users", wrapper(testMySQLListUsers, mySQLStore, mock))
	t.Run("mysql delete user", wrapper(testMySQLDeleteUser, mySQLStore, mock))

	// Rate limit
	t.Run("mysql get rate limit", wrapper(testMySQLGetRateLimits, mySQLStore, mock))
	t.Run("mysql put rate limit", wrapper(testMySQLPutRateLimits, mySQLStore, mock))
	t.Run("mysql delete rate limit", wrapper(testMySQLDeleteRateLimit, mySQLStore, mock))

	// Miner
	t.Run("mysql has miner", wrapper(testMySQLHasMiner, mySQLStore, mock))
	t.Run("mysql get user by miner", wrapper(testMySQLGetUserByMiner, mySQLStore, mock))
	t.Run("mysql list miners", wrapper(testMySQLListMiner, mySQLStore, mock))
	t.Run("mysql delete miner", wrapper(testMySQLDeleteMiner, mySQLStore, mock))
	t.Run("mysql upsert miner", wrapper(testMySQLUpsertMiner, mySQLStore, mock))

	// Signer
	t.Run("mysql has signer", wrapper(testMySQLHasSigner, mySQLStore, mock))
	t.Run("mysql get user by signer", wrapper(testMySQLGetUserBySigner, mySQLStore, mock))
	t.Run("mysql list signers", wrapper(testMySQLListSigner, mySQLStore, mock))
	t.Run("mysql delete signer", wrapper(testMySQLDeleteSigner, mySQLStore, mock))
	t.Run("mysql upsert signer", wrapper(testMySQLUpsertSigner, mySQLStore, mock))

	// Version
	t.Run("mysql get version", wrapper(testMySQLVersion, mySQLStore, mock))
	t.Run("mysql migrate to v1", wrapper(testMySQLMigrateToV1, mySQLStore, mock))

	if err = mysqlShutdown(mock, sqlDB); err != nil {
		t.Fatal(err)
	}
}

func testMySQLPutToken(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	kp := &KeyPair{
		Name:       "test_token_001",
		Perm:       "admin",
		Secret:     "d6234bf3f14a568a9c8315a6ee4f474e380beb2b65a64e6ba0142df72b454f4e",
		Extra:      "",
		Token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwtdG9rZW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.DQ-ETWoEnNrpGKCikwZax6YUzdQIkhT0pHOTSta8770",
		CreateTime: time.Now(),
		IsDeleted:  0,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `token` (`name`,`perm`,`secret`,`extra`,`token`,`createTime`,`is_deleted`) VALUES (?,?,?,?,?,?,?)")).
		WithArgs(kp.Name, kp.Perm, kp.Secret, kp.Extra, kp.Token, kp.CreateTime, kp.IsDeleted).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.Put(kp)
	assert.Nil(t, err)
}

func testMySQLUpdateToken(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	kp := &KeyPair{
		Name:       "test_token_001",
		Perm:       "admin",
		Secret:     "d6234bf3f14a568a9c8315a6ee4f474e380beb2b65a64e6ba0142df72b454f4e",
		Extra:      "",
		Token:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemwtdG9rZW4iLCJwZXJtIjoiYWRtaW4iLCJleHQiOiIifQ.DQ-ETWoEnNrpGKCikwZax6YUzdQIkhT0pHOTSta8770",
		CreateTime: time.Now(),
		IsDeleted:  0,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `token` SET `createTime`=?,`extra`=?,`is_deleted`=?,`name`=?,`perm`=?,`secret`=?,`token`=? WHERE token = ?")).
		WithArgs(kp.CreateTime, kp.Extra, kp.IsDeleted, kp.Name, kp.Perm, kp.Secret, kp.Token, kp.Token).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.UpdateToken(kp)
	assert.Nil(t, err)
}

func testMySQLHasToken(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	token := Token("test_token")

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `token` WHERE token=? and is_deleted=?")).
		WithArgs(token, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	has, err := mySQLStore.Has(token)
	assert.Nil(t, err)
	assert.True(t, has)
}

func testMySQLGetToken(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	name := "test_token_name"
	token := Token("test_token")

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `token` WHERE token = ? and is_deleted=? LIMIT 1")).
		WithArgs(token, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"name", "token"}).AddRow(name, token))

	tokenInfo, err := mySQLStore.Get(token)
	assert.Nil(t, err)
	assert.Equal(t, name, tokenInfo.Name)
	assert.Equal(t, token, tokenInfo.Token)
}

func testMySQLGetTokenRecord(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	name := "test_token_name"
	token := Token("test_token")

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `token` WHERE token = ? LIMIT 1")).
		WithArgs(token).
		WillReturnRows(sqlmock.NewRows([]string{"name", "token"}).AddRow(name, token))

	tokenInfo, err := mySQLStore.GetTokenRecord(token)
	assert.Nil(t, err)
	assert.Equal(t, name, tokenInfo.Name)
	assert.Equal(t, token, tokenInfo.Token)
}

func testMySQLGetTokenByName(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	name := "test_token_name"
	token := Token("test_token")

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `token` WHERE name = ? and is_deleted=?")).
		WithArgs(name, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"name", "token"}).AddRow(name, token))

	tokenInfo, err := mySQLStore.ByName(name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tokenInfo))
	assert.Equal(t, name, tokenInfo[0].Name)
	assert.Equal(t, token, tokenInfo[0].Token)
}

func testMySQLListTokens(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	var skip int64 = 2
	var limit int64 = 10

	mock.ExpectQuery(regexp.QuoteMeta(
		fmt.Sprintf("SELECT * FROM `token` WHERE is_deleted=? ORDER BY name LIMIT %v OFFSET %v", limit, skip))).
		WithArgs(core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"name", "token"}).AddRow("", ""))

	tokens, err := mySQLStore.List(skip, limit)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tokens))
}

func testMySQLDeleteToken(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	token := Token("test_token")

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `token` WHERE token=? and is_deleted=?")).
		WithArgs(token, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `token` SET `is_deleted`=? WHERE token=?")).
		WithArgs(core.Deleted, token).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.Delete(token)
	assert.Nil(t, err)
}

func testMySQLPutUser(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	now := time.Now()
	id := uuid.NewString()
	user := "test_user_001"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `users` (`id`,`name`,`comment`,`state`,`createTime`,`updateTime`,`is_deleted`) VALUES (?,?,?,?,?,?,?)")).
		WithArgs(id, user, "", 0, now, now, 0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.PutUser(&User{
		Id:         id,
		Name:       user,
		UpdateTime: now,
		CreateTime: now,
	})
	assert.Nil(t, err)
}

func testMySQLUpdateUser(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	now := time.Now()
	id := uuid.NewString()
	user := "test_user_001"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `users` SET `name`=?,`comment`=?,`state`=?,`createTime`=?,`updateTime`=?,`is_deleted`=? WHERE `id` = ?")).
		WithArgs(user, "", 0, now, now, 0, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.UpdateUser(&User{
		Id:         id,
		Name:       user,
		UpdateTime: now,
		CreateTime: now,
	})
	assert.Nil(t, err)
}

func testMySQLHasUser(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	user := "test_user_001"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `users` WHERE name=? and is_deleted=?")).
		WithArgs(user, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exist, err := mySQLStore.HasUser(user)
	assert.Nil(t, err)
	assert.True(t, exist)
}

func testMySQLGetUser(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	user := "test_user_001"
	comment := "comment"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `users` WHERE name=? and is_deleted=? LIMIT 1")).
		WithArgs(user, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"name", "comment"}).AddRow(user, comment))

	userInfo, err := mySQLStore.GetUser(user)
	assert.Nil(t, err)
	assert.Equal(t, userInfo.Name, user)
	assert.Equal(t, userInfo.Comment, comment)
}

func testMySQLGetUserRecord(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	user := "test_user_001"
	comment := "comment"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `users` WHERE name=? LIMIT 1")).
		WithArgs(user).
		WillReturnRows(sqlmock.NewRows([]string{"name", "comment"}).AddRow(user, comment))

	userInfo, err := mySQLStore.GetUserRecord(user)
	assert.Nil(t, err)
	assert.Equal(t, userInfo.Name, user)
	assert.Equal(t, userInfo.Comment, comment)
}

func testMySQLListUsers(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	var skip int64 = 2
	var limit int64 = 10

	mock.ExpectQuery(regexp.QuoteMeta(
		fmt.Sprintf("SELECT * FROM `users` WHERE is_deleted=? ORDER BY createTime LIMIT %v OFFSET %v", limit, skip))).
		WithArgs(core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("user1").AddRow("user2"))

	users, err := mySQLStore.ListUsers(skip, limit, core.UserStateUndefined)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
}

func testMySQLDeleteUser(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	user := "test_user_001"
	// addr, _ := address.NewFromString("f01222345678999")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `users` WHERE name=? and is_deleted=? LIMIT 1")).
		WithArgs(user, core.NotDelete).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(user))

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `miners` WHERE user = ? AND `miners`.`deleted_at` IS NULL")).
		WithArgs(user).
		// TODO: name "miner": non-string types unsupported
		WillReturnRows(sqlmock.NewRows([]string{"aa"}).AddRow("f01222345678999"))

	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `miners` SET `deleted_at`=? WHERE miner = ? AND `miners`.`deleted_at` IS NULL")).
		WithArgs(anyTime{}, "<empty>").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(""+
		"INSERT INTO `users` (`id`,`name`,`comment`,`state`,`createTime`,`updateTime`,`is_deleted`) VALUES (?,?,?,?,?,?,?)")).
		WithArgs("", user, "", 0, anyTime{}, anyTime{}, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := mySQLStore.DeleteUser(user)
	assert.Nil(t, err)
}

func testMySQLGetRateLimits(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	name := "name"
	id := "id"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `user_rate_limits` WHERE name = ? AND id = ?")).
		WithArgs(name, id).
		WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow(name, id))

	rateLimits, err := mySQLStore.GetRateLimits(name, id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rateLimits))
	assert.Equal(t, name, rateLimits[0].Name)
	assert.Equal(t, id, rateLimits[0].Id)
}

func testMySQLPutRateLimits(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	rateLimit := UserRateLimit{
		Id:      "id",
		Name:    "name",
		Service: "service",
		API:     "",
		ReqLimit: ReqLimit{
			Cap:      1,
			ResetDur: 10,
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `user_rate_limits` SET `name`=?,`service`=?,`api`=?,`reqLimit`=? WHERE `id` = ?")).
		WithArgs(rateLimit.Name, rateLimit.Service, rateLimit.API, rateLimit.ReqLimit, rateLimit.Id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := mySQLStore.PutRateLimit(&rateLimit)
	assert.Nil(t, err)
	assert.Equal(t, rateLimit.Id, id)
}

func testMySQLDeleteRateLimit(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	name := "name"
	id := "id"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM `user_rate_limits` WHERE id = ? and name= ?")).
		WithArgs(id, name).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.DelRateLimit(name, id)
	assert.Nil(t, err)
}

func testMySQLHasMiner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("f01000")
	assert.Nil(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `miners` WHERE miner = ? AND deleted_at IS NULL")).
		WithArgs(storedAddress(addr)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exist, err := mySQLStore.HasMiner(addr, "")
	assert.Nil(t, err)
	assert.True(t, exist)
}

func testMySQLGetUserByMiner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("f01000")
	assert.Nil(t, err)
	userName := "name"
	userId := "id"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT users.* FROM `miners` inner join users on miners.miner = ? and users.name = miners.user WHERE `miners`.`deleted_at` IS NULL")).
		WithArgs(storedAddress(addr)).
		WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow(userName, userId))

	user, err := mySQLStore.GetUserByMiner(addr)
	assert.Nil(t, err)
	assert.Equal(t, userName, user.Name)
	assert.Equal(t, userId, user.Id)
}

func testMySQLListMiner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	userName := "user_name"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `miners` WHERE user = ? AND `miners`.`deleted_at` IS NULL")).
		WithArgs(userName).
		WillReturnRows(sqlmock.NewRows([]string{"user"}).
			AddRow(userName).AddRow(userName))

	miners, err := mySQLStore.ListMiners(userName)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(miners))
	assert.Equal(t, userName, miners[0].User)
}

func testMySQLDeleteMiner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("f01000")
	assert.Nil(t, err)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `miners` SET `deleted_at`=? WHERE miner = ? AND `miners`.`deleted_at` IS NULL")).
		WithArgs(anyTime{}, storedAddress(addr)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	success, err := mySQLStore.DelMiner(addr)
	assert.Nil(t, err)
	assert.True(t, success)
}

func testMySQLUpsertMiner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("f01000")
	assert.Nil(t, err)
	user := "user"
	openMining := true

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `users` WHERE name = ? ORDER BY `users`.`id` LIMIT 1")).
		WithArgs(user).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(user))

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `miners` WHERE miner = ? AND `miners`.`deleted_at` IS NULL")).
		WithArgs(storedAddress(addr)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `miners` (`miner`,`user`,`open_mining`,`created_at`,`updated_at`,`deleted_at`) "+
			"VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `user`=VALUES(`user`),`open_mining`=VALUES(`open_mining`),"+
			"`updated_at`=VALUES(`updated_at`),`deleted_at`=VALUES(`deleted_at`)")).
		WithArgs(storedAddress(addr), user, openMining, anyTime{}, anyTime{}, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	isCreate, err := mySQLStore.UpsertMiner(addr, user, openMining)
	assert.Nil(t, err)
	assert.False(t, isCreate)
}

func testMySQLHasSigner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha")
	assert.Nil(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `signers` WHERE `signer` = ? AND `deleted_at` IS NULL")).
		WithArgs(storedAddress(addr)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exist, err := mySQLStore.HasSigner(addr, "")
	assert.Nil(t, err)
	assert.True(t, exist)
}

func testMySQLGetUserBySigner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha")
	assert.Nil(t, err)
	userName := "name"
	userId := "id"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT users.* FROM `signers` inner join users on signers.signer = ? and users.name = signers.user WHERE `signers`.`deleted_at` IS NULL")).
		WithArgs(storedAddress(addr)).
		WillReturnRows(sqlmock.NewRows([]string{"name", "id"}).AddRow(userName, userId))

	user, err := mySQLStore.GetUserBySigner(addr)
	assert.Nil(t, err)
	assert.Equal(t, userName, user.Name)
	assert.Equal(t, userId, user.Id)
}

func testMySQLListSigner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	userName := "user_name"

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `signers` WHERE user = ? AND `signers`.`deleted_at` IS NULL")).
		WithArgs(userName).
		WillReturnRows(sqlmock.NewRows([]string{"user"}).
			AddRow(userName).AddRow(userName))

	signers, err := mySQLStore.ListSigner(userName)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(signers))
	assert.Equal(t, userName, signers[0].User)
}

func testMySQLDeleteSigner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha")
	assert.Nil(t, err)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE `signers` SET `deleted_at`=? WHERE `signer` = ? AND `signers`.`deleted_at` IS NULL")).
		WithArgs(anyTime{}, storedAddress(addr)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	success, err := mySQLStore.DelSigner(addr)
	assert.Nil(t, err)
	assert.True(t, success)
}

func testMySQLUpsertSigner(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	addr, err := address.NewFromString("t3wylwd6pclppme4qmbgwled5xpsbgwgqbn2alxa7yahg2gnbfkipsdv6m764xm5coizujmwdmkxeugplmorha")
	assert.Nil(t, err)
	user := "user"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `users` WHERE `name` = ? ORDER BY `users`.`id` LIMIT 1")).
		WithArgs(user).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(user))

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT count(*) FROM `signers` WHERE `signer` = ? AND `signers`.`deleted_at` IS NULL")).
		WithArgs(storedAddress(addr)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `signers` (`signer`,`user`,`created_at`,`updated_at`,`deleted_at`) "+
			"VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE `user`=VALUES(`user`),`updated_at`=VALUES(`updated_at`),"+
			"`deleted_at`=VALUES(`deleted_at`)")).
		WithArgs(storedAddress(addr), user, anyTime{}, anyTime{}, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	isCreate, err := mySQLStore.UpsertSigner(addr, user)
	assert.Nil(t, err)
	assert.False(t, isCreate)
}

func testMySQLVersion(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {
	correctVersion := uint64(3)

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `store_versions` ORDER BY `store_versions`.`id` LIMIT 1")).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(correctVersion))

	version, err := mySQLStore.Version()
	assert.Nil(t, err)
	assert.Equal(t, correctVersion, version)
}

func testMySQLMigrateToV1(t *testing.T, mySQLStore *mysqlStore, mock sqlmock.Sqlmock) {

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT * FROM `users`")).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("name"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(
		"INSERT INTO `store_versions` (`version`,`id`) VALUES (?,?) ON DUPLICATE KEY UPDATE `version`=VALUES(`version`)")).
		WithArgs(1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := mySQLStore.MigrateToV1()
	assert.Nil(t, err)
}

func mysqlSetup() (*mysqlStore, sqlmock.Sqlmock, *sql.DB, error) {
	var err error
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, nil, err
	}

	mock.ExpectQuery("SELECT VERSION()").WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(""))

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	mySQLStore := &mysqlStore{db: gormDB}
	return mySQLStore, mock, sqlDB, err
}

func mysqlShutdown(mock sqlmock.Sqlmock, sqlDB *sql.DB) error {
	mock.ExpectClose()
	return sqlDB.Close()
}
