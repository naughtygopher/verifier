package stores

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack/v4"

	"github.com/naughtygopher/verifier"
)

// RedisConfig holds all the configuration required for the redis handler
type RedisConfig struct {
	Hosts        []string      `json:"hosts,omitempty"`
	Username     string        `json:"username,omitempty"`
	Password     string        `json:"password,omitempty"`
	DialTimeout  time.Duration `json:"dialTimeoutSecs,omitempty"`
	ReadTimeout  time.Duration `json:"readTimeoutSecs,omitempty"`
	WriteTimeout time.Duration `json:"writeTimeoutSecs,omitempty"`
}

// Redis struct exposes all the store functionalities required for verifier
type Redis struct {
	client redis.UniversalClient
}

// Create creates a new entry of the verification request in the store
func (ris *Redis) Create(ver *verifier.Request) (*verifier.Request, error) {
	key := fmt.Sprintf(
		"%s-%s",
		ver.Type,
		ver.Recipient,
	)

	now := time.Now()
	expiry := ver.SecretExpiry.Sub(now)

	payload, err := msgpack.Marshal(ver)
	if err != nil {
		return nil, err
	}

	resp := ris.client.Set(
		key,
		payload,
		expiry,
	)

	err = resp.Err()
	if err != nil {
		return nil, err
	}

	return ver, nil
}

// ReadLastPending reads the last pending verification request of the commtype + recipient
func (ris *Redis) ReadLastPending(ctype verifier.CommType, recipient string) (*verifier.Request, error) {
	key := fmt.Sprintf(
		"%s-%s",
		ctype,
		recipient,
	)

	resp := ris.client.Get(key)
	err := resp.Err()
	if err != nil {
		return nil, err
	}

	b, err := resp.Bytes()
	if err != nil {
		return nil, err
	}

	ver := &verifier.Request{}
	err = msgpack.Unmarshal(b, ver)
	if err != nil {
		return nil, err
	}

	return ver, nil
}

// Update updates a verification request for the given verification ID & the payload
func (ris *Redis) Update(verID string, ver *verifier.Request) (*verifier.Request, error) {
	key := fmt.Sprintf(
		"%s-%s",
		ver.Type,
		ver.Recipient,
	)
	now := time.Now()
	expiry := ver.SecretExpiry.Sub(now)

	payload, err := msgpack.Marshal(ver)
	if err != nil {
		return nil, err
	}

	resp := ris.client.Set(
		key,
		payload,
		expiry,
	)

	err = resp.Err()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// NewRedis returns a newly initialized redis store
func NewRedis(cfg *RedisConfig) (*Redis, error) {
	cli := redis.NewUniversalClient(
		&redis.UniversalOptions{
			Addrs:        cfg.Hosts,
			Password:     cfg.Password,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			PoolTimeout:  cfg.WriteTimeout * 10,
		},
	)

	result := cli.Ping()
	err := result.Err()
	if err != nil {
		return nil, err
	}

	r := &Redis{
		client: cli,
	}
	return r, nil
}
