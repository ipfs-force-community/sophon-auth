package core

var CurrentCommit string

const BuildVersion = "1.11.0"

var Version = BuildVersion + CurrentCommit

const EmptyString = ""

const AuthorizationHeader = "Authorization"

type (
	DBPrefix   = []byte
	Permission = string
)

const (
	// When changing these, update docs/API.md too
	PermRead  Permission = "read" // default
	PermWrite Permission = "write"
	PermSign  Permission = "sign"  // Use wallet keys for signing
	PermAdmin Permission = "admin" // Manage permissions
)

var PermArr = []Permission{
	PermRead, PermWrite, PermSign, PermAdmin,
}

func IsValid(perm Permission) bool {
	for _, v := range PermArr {
		if v == perm {
			return true
		}
	}
	return false
}

func AdaptOldStrategy(perm Permission) []Permission {
	perms := make([]Permission, 0, 4)
	switch perm {
	case PermAdmin:
		perms = append(perms, PermRead, PermWrite, PermSign, PermAdmin)
	case PermSign:
		perms = append(perms, PermRead, PermWrite, PermSign)
	case PermWrite:
		perms = append(perms, PermRead, PermWrite)
	case PermRead:
		perms = append(perms, PermRead)
	default:
	}
	return perms
}

type (
	LogField    = string
	Measurement = string
)

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

// For compatibility with older versions of the API
const (
	UserStateUndefined UserState = iota
	UserStateEnabled
	UserStateDisabled
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
