package storage

type TempStoreService interface {
	Exist(key string) (bool, error)
	Get(key string) ([]byte, error)
	GetAllByPattern(pattern string) ([][]byte, error)
	GetAllByKeys(keys ...string) ([][]byte, error)
	Set(key string, response []byte, ttl int64) error
	Del(key string) error
	GetKeyIncrement(key string) (int64, error)
	Incr(key string) error
}
