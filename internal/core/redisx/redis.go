package redisx

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"strconv"
	"time"
)

var Nil = redis.Nil

type Redis struct {
	rdsClient *redis.Client
}

func New() *Redis {
	return &Redis{}
}
func (r *Redis) getRedis() (*redis.Client, error) {
	if r.rdsClient != nil {
		return r.rdsClient, nil
	}
	// 初始化RedisClient
	r.rdsClient = redis.NewClient(&redis.Options{
		Addr:        viper.GetString("redis.addr"),
		DB:          viper.GetInt("redis.database"),
		Password:    viper.GetString("redis.password"),
		PoolSize:    viper.GetInt("redis.pool.size"),
		PoolTimeout: viper.GetDuration("redis.pool.timeout"),
	})
	_, err := r.rdsClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return r.rdsClient, nil
}

func (r *Redis) Get(key string) (string, error) {
	conn, err := r.getRedis()
	if err != nil {
		return "", err
	}
	value, err := conn.Get(context.Background(), key).Result()
	if err == Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *Redis) Set(key string, value string, exp time.Duration) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.Set(context.Background(), key, value, exp).Err()
}

func (r *Redis) Del(key string) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.Del(context.Background(), key).Err()
}

func (r *Redis) HSet(key string, value ...interface{}) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.HSet(context.Background(), key, value...).Err()
}

func (r *Redis) HGetAll(key string) (map[string]string, error) {
	conn, err := r.getRedis()
	if err != nil {
		return nil, err
	}
	return conn.HGetAll(context.Background(), key).Result()
}

func (r *Redis) HGet(key string, hash string) (string, error) {
	conn, err := r.getRedis()
	if err != nil {
		return "", err
	}
	val, err := conn.HGet(context.Background(), key, hash).Result()
	if err == Nil {
		return "", nil
	}
	return val, err
}

func (r *Redis) HExists(key string, hash string) (bool, error) {
	conn, err := r.getRedis()
	if err != nil {
		return false, err
	}
	return conn.HExists(context.Background(), key, hash).Result()
}

func (r *Redis) HDel(key string, hash string) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.HDel(context.Background(), key, hash).Err()
}

func (r *Redis) HCount(key string) int64 {
	conn, err := r.getRedis()
	if err != nil {
		return 0
	}
	return conn.HLen(context.Background(), key).Val()
}

func (r *Redis) Expire(key string, ttl time.Duration) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.Expire(context.Background(), key, ttl).Err()
}

func (r *Redis) ZAdd(key string, value string, score int64) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.ZAdd(context.Background(), key, redis.Z{Score: float64(score), Member: value}).Err()
}

func (r *Redis) ZRem(key string, value string) error {
	conn, err := r.getRedis()
	if err != nil {
		return err
	}
	return conn.ZRem(context.Background(), key, value).Err()
}

func (r *Redis) ZRangeByScore(key string, start, stop int64) ([]string, error) {
	conn, err := r.getRedis()
	if err != nil {
		return nil, err
	}
	return conn.ZRangeByScore(context.Background(), key, &redis.ZRangeBy{
		Min: strconv.FormatInt(start, 10),
		Max: strconv.FormatInt(stop, 10),
	}).Result()
}
