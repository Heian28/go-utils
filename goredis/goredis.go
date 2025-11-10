package goredis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Heian28/go-utils/gologger"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// GoRedisConfig is a struct that represents the configuration for the GoRedis.
// It is used to represent the configuration for the GoRedis.
type GoRedisConfig struct {
	Addr     string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// GoRedis is an interface that defines the methods for the GoRedis.
// It is used to define the methods for the GoRedis.
type GoRedis interface {
	Save(
		ctx context.Context,
		key string,
		value any,
		ttl time.Duration,
	) error
	Get(
		ctx context.Context,
		key string,
		dest any,
	) error
	Delete(
		ctx context.Context,
		key string,
	) error
	DeleteByPattern(
		ctx context.Context,
		pattern string,
		batch int64,
	) error
}

// rds is a struct that represents the redis.
// It is used to represent the redis.
type rds struct {
	rdb       *redis.Client
	conf      GoRedisConfig
	log       *logrus.Logger
	withDebug bool
}

// New is a function that creates a new GoRedis.
// It takes a GoRedisConfig, a pointer to a logrus.Logger, and a bool and returns a GoRedis.
// This is used to create a new GoRedis.
func New(conf GoRedisConfig, log *logrus.Logger, withDebug bool) GoRedis {
	if log == nil {
		gologger.New(
			gologger.SetIsProduction(true),
			gologger.SetServiceName("GoRedis"),
		)
		log = gologger.Logger
	}
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Addr,
		Password:     conf.Password,
		DB:           conf.Database,
		PoolSize:     10,
		MinIdleConns: 10,
		MaxRetries:   3,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolTimeout:  10 * time.Second,
	})

	log.Infof("[GoRedis] Connected to Redis %s...", fmt.Sprintf("redis://%s/%d", conf.Addr, conf.Database))

	return &rds{
		rdb:       client,
		conf:      conf,
		log:       log,
		withDebug: withDebug,
	}
}

// Save is a function that saves the value to the redis.
// It takes a context, a string, a any, and a time.Duration and returns an error.
// This is used to save the value to the redis.
func (r *rds) Save(ctx context.Context, key string, value any, ttl time.Duration) error {
	r.log.Infof("[GoRedis] Saving to Redis %s...", key)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if r.withDebug {
		r.log.Debug(string(data))
	}
	return r.rdb.Set(ctx, key, data, ttl).Err()
}

// Get is a function that gets the value from the redis.
// It takes a context, a string, and a pointer to a any and returns an error.
// This is used to get the value from the redis.
func (r *rds) Get(ctx context.Context, key string, dest any) error {
	r.log.Infof("[GoRedis] Getting from Redis %s...", key)
	data, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if r.withDebug {
		r.log.Debug(data)
	}
	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return err
	}
	return nil
}

// Delete is a function that deletes the value from the redis.
// It takes a context, a string, and returns an error.
// This is used to delete the value from the redis.
func (r *rds) Delete(ctx context.Context, key string) error {
	r.log.Infof("[GoRedis] Deleting from Redis %s...", key)
	return r.rdb.Del(ctx, key).Err()
}

// DeleteByPattern is a function that deletes the value from the redis using a pattern.
// It takes a context, a string, and a int64 and returns an error.
// This is used to delete the value from the redis using a pattern.
func (r *rds) DeleteByPattern(ctx context.Context, pattern string, batch int64) error {
	r.log.Infof("[GoRedis] Deleting using pattern from Redis %s...", pattern)
	var cursor uint64
	if batch == 0 {
		batch = 100
	}
	for {
		k, nc, err := r.rdb.Scan(ctx, cursor, pattern, batch).Result()
		if err != nil {
			return err
		}
		if len(k) > 0 {
			if err := r.rdb.Del(ctx, k...).Err(); err != nil {
				return err
			}
		}
		cursor = nc
		if cursor == 0 {
			break
		}
	}
	return nil
}
