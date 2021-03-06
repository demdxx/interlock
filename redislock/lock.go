// Package redislock implements locker based on Redis server
package redislock

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/demdxx/gocast"
	"github.com/go-redis/redis"
)

var errLockHasFailed = errors.New(`redis lock has failed`)

// Lock provides redis key locker
type Lock struct {
	lifetime time.Duration
	client   redis.Cmdable
}

// New returns redis Lock for redis client
func New(client redis.Cmdable, defaultLifetime time.Duration) *Lock {
	return &Lock{client: client, lifetime: defaultLifetime}
}

// NewByURL returns redis Lock object or error
func NewByURL(connectURL string, defaultLifetime time.Duration) (*Lock, error) {
	var (
		connectURLObj, err = url.Parse(connectURL)
		password           string
	)
	if err != nil {
		return nil, err
	}
	if connectURLObj.User != nil {
		password, _ = connectURLObj.User.Password()
	}
	return New(redis.NewClient(&redis.Options{
		DB:           gocast.ToInt(strings.Trim(connectURLObj.Path, `/`)),
		Addr:         connectURLObj.Host,
		Password:     password,
		PoolSize:     gocast.ToInt(connectURLObj.Query().Get(`pool`)),
		MaxRetries:   gocast.ToInt(connectURLObj.Query().Get(`max_retries`)),
		MinIdleConns: gocast.ToInt(connectURLObj.Query().Get(`idle_cons`)),
	}), defaultLifetime), nil
}

// TryLock message as processing
func (mr *Lock) TryLock(key interface{}, lifetime ...time.Duration) error {
	lt := mr.lifetime
	if len(lifetime) == 1 {
		lt = lifetime[0]
	}
	res, err := mr.client.SetNX(hash(key), []byte(`t`), lt).Result()
	if err == nil && !res {
		err = errLockHasFailed
	}
	return err
}

// IsLocked in the redis server
func (mr *Lock) IsLocked(key interface{}) bool {
	val, _ := mr.client.Get(hash(key)).Result()
	return val == `t`
}

// Unlock message as processing
func (mr *Lock) Unlock(key interface{}) error {
	return mr.client.Del(hash(key)).Err()
}
