package storage

import (
	"errors"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type redisStoreService struct {
	redisPool *redis.Pool
}

func (t *redisStoreService) Exist(key string) (bool, error) {
	conn := t.redisPool.Get()
	defer conn.Close()

	return redis.Bool(conn.Do("EXISTS", key))
}

func (t *redisStoreService) Get(key string) ([]byte, error) {
	conn := t.redisPool.Get()
	defer conn.Close()

	result, err := redis.Bytes(conn.Do("GET", key))
	if err == redis.ErrNil {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("redisStoreService.Get, conn.Do GET key error: %s", err)
	}

	return result, nil
}

func (t *redisStoreService) GetAllByPattern(pattern string) ([][]byte, error) {
	conn := t.redisPool.Get()
	defer conn.Close()

	var keyArgs []interface{}

	keyArgs, err := redis.Values(conn.Do("KEYS", pattern+"*"))
	if err != nil {
		return nil, fmt.Errorf(
			"redisStoreService.GetAllByPattern, conn.Do KEYS pattern %s error: %s", pattern, err)
	}

	values, err := redis.ByteSlices(conn.Do("MGET", keyArgs...))
	if err != nil {
		return nil, fmt.Errorf(
			"redisStoreService.GetAllByPattern, conn.Do MGET pattern %s error: %s", pattern, err)
	}

	return values, nil
}

func (t *redisStoreService) GetAllByKeys(keys ...string) ([][]byte, error) {
	conn := t.redisPool.Get()
	defer conn.Close()

	var keyArgs []interface{}

	for _, k := range keys {
		keyArgs = append(keyArgs, k)
	}

	values, err := redis.ByteSlices(conn.Do("MGET", keyArgs...))
	if err != nil {
		return nil, fmt.Errorf(
			"redisStoreService.GetAllByKeys, conn.Do MGET error: %s", err)
	}

	return values, nil
}

func (t *redisStoreService) Set(key string, resp []byte, ttl int64) error {
	conn := t.redisPool.Get()
	defer conn.Close()

	var err error
	if ttl > 0 {
		_, err = conn.Do("SETEX", key, ttl, resp)
	} else {
		_, err = conn.Do("SET", key, resp)
	}

	if err != nil {
		return fmt.Errorf("redisStoreService.Set, unable to execute SET/SETEX command: %s", err)
	}

	return nil
}

func (t *redisStoreService) Del(key string) error {
	conn := t.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return fmt.Errorf("redisStoreService.Del, unable to execute DEL command: %s", err)
	}

	return nil
}

func (t *redisStoreService) GetKeyIncrement(key string) (int64, error) {
	conn := t.redisPool.Get()
	defer conn.Close()

	result, err := redis.Int64(conn.Do("GET", key))
	if err == redis.ErrNil {
		// Set increment key, if it's not yet exist, starts from 1.
		const startsFrom = 1

		_, err = conn.Do("SET", key, startsFrom)
		if err != nil {
			return 0, fmt.Errorf(
				"redisStoreService.KeyIncrementedId, conn.Do SET new increment key error: %s", err)
		}

		return startsFrom, nil
	}

	if err != nil {
		return 0, fmt.Errorf("redisStoreService.KeyIncrementedId, conn.Do GET key error: %s", err)
	}

	return result, nil

}

func (t *redisStoreService) Incr(key string) error {
	conn := t.redisPool.Get()
	defer conn.Close()

	_, err := redis.Int64(conn.Do("INCR", key))
	if err != nil {
		return fmt.Errorf("redisStoreService.Incr, unable to execute INCR command: %s", err)
	}

	return nil
}

func NewRedisStoreService(redisPool *redis.Pool) (TempStoreService, error) {
	if redisPool == nil {
		return nil, errors.New("redis pool must be not empty")
	}

	return &redisStoreService{
		redisPool: redisPool,
	}, nil
}
