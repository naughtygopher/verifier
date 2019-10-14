package verifier

type store interface {
	Create(ver *Verification) (*Verification, error)
	ReadLastPending(ctype commType, recipient string) (*Verification, error)
	Update(verID string, ver *Verification) (*Verification, error)
}

type redisStore struct {
}

func (rs *redisStore) Create(ver *Verification) (*Verification, error) {
	return ver, nil
}

func (rs *redisStore) ReadLastPending(ctype commType, recipient string) (*Verification, error) {
	return nil, nil
}

func (rs *redisStore) Update(verID string, ver *Verification) (*Verification, error) {
	return ver, nil
}

func newstore() (*redisStore, error) {
	return &redisStore{}, nil
}
