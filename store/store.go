package store

type Store map[string]string

func New() Store {
	return make(Store)
}

type ErrorNoSuchKey struct {
	key string
}

func (e ErrorNoSuchKey) Error() string {
	return "no such key for key " + e.key
}

func (s Store) Put(key, value string) error {
	s[key] = value
	return nil
}

func (s Store) Get(key string) (string, error) {
	val, ok := s[key]
	if !ok {
		return val, ErrorNoSuchKey{key: key}
	}
	return val, nil
}

func (s Store) Delete(key string) error {
	delete(s, key)
	return nil
}
