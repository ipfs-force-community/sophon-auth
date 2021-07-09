package storage

type Prefix = string

const (
	PrefixToken   Prefix = "TOKEN:"
	PrefixAccount Prefix = "USER:"
)

func (s *badgerStore) accountKey(name string) []byte {
	return []byte(PrefixAccount + name)
}

func (s *badgerStore) tokenKey(name string) []byte {
	return []byte(PrefixToken + name)
}
