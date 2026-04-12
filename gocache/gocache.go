package gocache

import (
	"context"
	"fmt"
	"time"

	"github.com/Heian28/go-utils/gologger"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
)

type GoCacheConfig struct {
	URI          string        `mastructure:"uri"`
	User         string        `mastructure:"user"`
	Password     string        `mastructure:"password"`
	Database     int           `mastructure:"database"`
	PoolSize     int           `mastructure:"poolSize"`
	MinIdleConns int           `mastructure:"minIdleConns"`
	MaxRetries   int           `mastructure:"maxRetries"`
	ReadTimeout  time.Duration `mastructure:"readTimeout"`
	WriteTimeout time.Duration `mastructure:"writeTimeout"`
	PoolTimeout  time.Duration `mastructure:"poolTimeout"`
}

type GoCacheManager interface {
	Save(ctx context.Context, key string, value any, duration time.Duration) error
	Get(ctx context.Context, key string, output any) error
	Delete(ctx context.Context, key string) error
	GetByPattern(ctx context.Context, pattern string, output any) error
	DeleteByPattern(ctx context.Context, pattern string, batch int) error
	Upsert(ctx context.Context, key string, value any, duration *time.Duration) error
}

type cache struct {
	client *redis.Client
	conf   GoCacheConfig
	log    *logrus.Logger
}

func New(conf GoCacheConfig, log *logrus.Logger) GoCacheManager {
	log = getLogger(log)
	client := redis.NewClient(setup(conf))

	log.Infof(
		"[GoCacheManager] Connected to redis database %s...",
		fmt.Sprintf("redis://%s/%d", conf.URI, conf.Database),
	)

	return &cache{
		client: client,
		conf:   conf,
		log:    log,
	}
}

func (c *cache) Save(ctx context.Context, key string, value any, duration time.Duration) error {
	c.log.Infof("[GoCacheManager] Saving to Redis %s...", key)

	data, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	if err := c.client.Set(ctx, key, data, duration).Err(); err != nil {
		return err
	}

	c.log.Info("[GoCacheManager] Saved successfully")
	return nil
}

func (c *cache) Get(ctx context.Context, key string, output any) error {
	c.log.Infof("[GoCacheManager] Getting from Redis %s...", key)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	if err := msgpack.Unmarshal([]byte(data), output); err != nil {
		return err
	}

	c.log.Info("[GoCacheManager] Get successfully")
	return nil
}

func (c *cache) Delete(ctx context.Context, key string) error {
	c.log.Infof("[GoCacheManager] Deleting from Redis %s...", key)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	c.log.Info("[GoCacheManager] Deleted successfully")
	return nil
}

func (c *cache) GetByPattern(ctx context.Context, pattern string, output any) error {
	c.log.Infof("[GoCacheManager] Getting by pattern %s...", pattern)

	var keys []string
	var cursor uint64
	for {
		var batch []string
		batch, cursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return nil
	}

	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return err
	}

	data, err := msgpack.Marshal(values)
	if err != nil {
		return err
	}

	if err := msgpack.Unmarshal(data, output); err != nil {
		return err
	}

	c.log.Info("[GoCacheManager] GetByPattern successfully")
	return nil
}

func (c *cache) DeleteByPattern(ctx context.Context, pattern string, batch int) error {
	c.log.Infof("[GoCacheManager] Deleting by pattern %s...", pattern)

	var cursor uint64
	for {
		var keys []string
		keys, cursor, err := c.client.Scan(ctx, cursor, pattern, int64(batch)).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	c.log.Info("[GoCacheManager] DeleteByPattern successfully")
	return nil
}

func (c *cache) Upsert(ctx context.Context, key string, value any, duration *time.Duration) error {
	c.log.Infof("[GoCacheManager] Upserting to Redis %s...", key)

	data, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	if duration != nil {
		if err := c.client.Set(ctx, key, data, *duration).Err(); err != nil {
			return err
		}
	} else {
		ttl, err := c.client.TTL(ctx, key).Result()
		if err != nil {
			return err
		}

		if ttl > 0 {
			if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
				return err
			}
		} else {
			if err := c.client.Set(ctx, key, data, 0).Err(); err != nil {
				return err
			}
		}
	}

	c.log.Info("[GoCacheManager] Upserted successfully")
	return nil
}

func setup(conf GoCacheConfig) *redis.Options {
	opt := &redis.Options{
		Addr:         conf.URI,
		DB:           conf.Database,
		MaxRetries:   5,
		PoolSize:     10,
		MinIdleConns: 10,
		PoolTimeout:  10 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if conf.User != "" && conf.Password != "" {
		opt.Username = conf.User
		opt.Password = conf.Password
	}

	return opt
}

func getLogger(log *logrus.Logger) *logrus.Logger {
	if log == nil {
		gologger.New(
			gologger.SetServiceName("GoCacheManager"),
			gologger.SetIsProduction(true),
		)
		log = gologger.Logger
	}
	return log
}
