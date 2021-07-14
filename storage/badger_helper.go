package storage

type Prefix = string

const (
	PrefixToken    Prefix = "TOKEN:"
	PrefixUser     Prefix = "USER:"
	PrefixReqLimit Prefix = "ReqLimit:"
)

func (s *badgerStore) rateLimitKey(name string) []byte {
	return []byte(PrefixReqLimit + name)
}

func (s *badgerStore) userKey(name string) []byte {
	return []byte(PrefixUser + name)
}

func (s *badgerStore) tokenKey(name string) []byte {
	return []byte(PrefixToken + name)
}
