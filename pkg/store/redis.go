package store

import (
	"context"
	"github.com/go-redis/redis/v9"
	"github.com/spf13/viper"
)

var (
	client *redis.Client
	ctx    = context.Background()
)

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:        viper.GetString("redis.addr"),
		DB:          viper.GetInt("redis.database"),
		Password:    viper.GetString("redis.password"),
		PoolSize:    viper.GetInt("redis.pool.size"),
		PoolTimeout: viper.GetDuration("redis.pool.timeout"),
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}
