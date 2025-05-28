package limiter

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Storage interface {
	Increment(key string) (int, error)
	Block(key string, duration time.Duration) error
	IsBlocked(key string) (bool, error)
	Ping() error
}

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr, password, db string) *RedisStorage {
	dbIndex, _ := strconv.Atoi(db)
	if db == "" {
		dbIndex = 0
	}

	return &RedisStorage{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       dbIndex,
		}),
	}
}

func (r *RedisStorage) Increment(key string) (int, error) {
	val, err := r.client.Incr(context.Background(), key).Result()
	return int(val), err
}

func (r *RedisStorage) Block(key string, duration time.Duration) error {
	return r.client.Set(context.Background(), key, "blocked", duration).Err()
}

func (r *RedisStorage) IsBlocked(key string) (bool, error) {
	val, err := r.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "blocked", nil
}

func (r *RedisStorage) Ping() error {
	return r.client.Ping(context.Background()).Err()
}
