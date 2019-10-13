package verifier

import "database/sql"

type store interface {
	Create(ver *Verification) (*Verification, error)
	ReadLastPending(ctype commType, recipient string) (*Verification, error)
	Update(verID string, ver *Verification) (*Verification, error)
}

type verifierStore struct {
	tableName string
	driver    *sql.DB
}

func (vs *verifierStore) Create(ver *Verification) (*Verification, error) {
	return nil, nil
}

func (vs *verifierStore) ReadLastPending(ctype commType, recipient string) (*Verification, error) {
	return nil, nil
}

func (vs *verifierStore) Update(verID string, ver *Verification) (*Verification, error) {
	return nil, nil
}

func newstore(driver *sql.DB) (*verifierStore, error) {
	vs := &verifierStore{
		tableName: "VerifcationRequests",
		driver:    driver,
	}
	return vs, nil
}
