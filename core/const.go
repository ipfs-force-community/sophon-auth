package core

import (
	"context"
	"errors"
)

var CurrentCommit string

const BuildVersion = "1.8.0-rc3"

var Version = BuildVersion + CurrentCommit

const EmptyString = ""

type DBPrefix = []byte

var (
	PrefixNode = DBPrefix("node")
)

const (
	ServiceToken = "Authorization"
)

var (
	// net work name set by cli
	NameSpace string
)

type Permission = string

const (
	// When changing these, update docs/API.md too
	PermRead  Permission = "read" // default
	PermWrite Permission = "write"
	PermSign  Permission = "sign"  // Use wallet keys for signing
	PermAdmin Permission = "admin" // Manage permissions
)

var (
	PermArr = []Permission{
		PermAdmin, PermSign, PermWrite, PermRead,
	}
)
var ErrPermIllegal = errors.New("perm illegal")

func ContainsPerm(perm Permission) error {
	for _, v := range PermArr {
		if v == perm {
			return nil
		}
	}
	return ErrPermIllegal
}

func AdaptOldStrategy(perm Permission) []Permission {
	perms := make([]Permission, 0)
	switch perm {
	case PermAdmin:
		perms = append(perms, PermAdmin, PermSign, PermWrite, PermRead)
	case PermSign:
		perms = append(perms, PermSign, PermWrite, PermRead)
	case PermWrite:
		perms = append(perms, PermWrite, PermRead)
	case PermRead:
		perms = append(perms, PermRead)
	default:
	}
	return perms
}

type PermKey int

var PermCtxKey PermKey

func WithPerm(ctx context.Context, perm Permission) context.Context {
	return context.WithValue(ctx, PermCtxKey, AdaptOldStrategy(perm))
}

type LogField = string
type Measurement = string

const (
	MTMethod Measurement = "method"
)
const (
	FieldName    LogField = "name"
	FieldIP      LogField = "ip"
	FieldLevel   LogField = "level"
	FieldSvcName LogField = "svcName"
	FieldSpanId  LogField = "spanId"
	FieldPreHost LogField = "preHost"
	FieldElapsed LogField = "elapsed"
	FieldToken   LogField = "token"
)

var TagFields = []LogField{
	FieldName,
	FieldIP,
	FieldLevel,
	FieldSvcName,
}

// request params code sum,enum 1 2 4 8, to multi-select
type KeyCode = int
type SourceType = int

const (
	Miner SourceType = 1
)

type Page struct {
	Skip  int64 `form:"skip" json:"skip"`
	Limit int64 `form:"limit" json:"limit"`
}

func (o *Page) GetSkip() int64 {
	if o.Skip < 0 {
		o.Skip = 0
	}
	return o.Skip
}
func (o *Page) GetLimit() int64 {
	if o.Limit < 0 || o.Limit > 1000 {
		o.Limit = 1000
	}
	return o.Limit
}

type UserState int

const (
	UserStateDisabled UserState = 0
	UserStateEnabled  UserState = 1
)

var userStateStrs = map[UserState]string{
	UserStateEnabled:  "enabled",
	UserStateDisabled: "disabled",
}

func (us UserState) String() string {
	stateStr, exists := userStateStrs[us]
	if exists {
		return stateStr
	}
	return "unknown state"
}

const (
	NotDelete = 0
	Deleted   = 1
)
