package services

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"time"

	"github.com/apsdehal/go-logger"
	"github.com/garyburd/redigo/redis"
)

const (
	redisHost = "redis"
	redisPort = "6379"

	redisMaxIdle     = 3
	redisIdleTimeout = 240 * time.Second
)

type App interface {
	Init() error
	RedisPool() *redis.Pool
	Logger() logger.Logger
}

type app struct {
	redisPool *redis.Pool
	logger    logger.Logger
}

func (app *app) RedisPool() *redis.Pool {
	return app.redisPool
}

func (app *app) Logger() logger.Logger {
	return app.logger
}

func (app *app) Init() error {
	redisPool, err := app.initRedisPool()
	if err != nil {
		return err
	}
	app.redisPool = redisPool

	logger, err := app.initLogger()
	if err != nil {
		return err
	}
	app.logger = *logger

	return nil
}

func (app *app) initRedisPool() (*redis.Pool, error) {
	if redisHost == "" || redisPort == "" {
		return nil, errors.New("initRedisPool init fail, check redisHost or redisPort")
	}
	address := fmt.Sprintf("%s:%s", redisHost, redisPort)

	// Init redis pool.
	p := &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}

	return p, nil
}

func (app *app) initLogger() (*logger.Logger, error) {
	log, err := logger.New("common", 1, os.Stdout)
	if err != nil {
		return nil, err
	}

	return log, nil
}

var instanceApp App

// Get initialized application or initialize and return it
func Application() App {
	if instanceApp == nil {
		instanceApp = new(app)
	}

	return instanceApp
}
