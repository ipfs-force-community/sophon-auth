package core

import "context"

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
