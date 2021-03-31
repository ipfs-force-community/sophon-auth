package core

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
