package storage

type Prefix = string

const (
	PrefixUser Prefix = "UR:"
)

func (s *badgerStore) userKey(name string) []byte {
	return []byte(PrefixUser + name)
}
