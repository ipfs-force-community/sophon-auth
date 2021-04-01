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
		break
	case PermSign:
		perms = append(perms, PermSign, PermWrite, PermRead)
		break
	case PermWrite:
		perms = append(perms, PermWrite, PermRead)
		break
	case PermRead:
		perms = append(perms, PermRead)
		break
	}
	return perms
}

type permKey int

var permCtxKey permKey

func WithPerm(ctx context.Context, perms []Permission) context.Context {
	return context.WithValue(ctx, permCtxKey, perms)
}

type LogField = string
type Measurement = string

const (
	MTMethod Measurement = "method"
)
const (
	FieldName  LogField = "name"
	FieldIP    LogField = "ip"
	FieldLevel LogField = "level"
)

var LogFields = []LogField{
	FieldName,
	FieldIP,
	FieldLevel,
}
