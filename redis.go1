package main

import (
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

const host = "host"
const prevHost = "prev_host"
const redisFallbackAddress = "localhost:6379"

type RedisClient struct {
	*redis.Client
}

var redisClient *RedisClient
var once sync.Once

const key = "drivers"

// GetRedisClient get Redis client
func GetRedisClient() *RedisClient {
	once.Do(func() {
		redisAddress := getEnv("REDIS_ADDRESS", redisFallbackAddress)
		opt, _ := redis.ParseURL(redisAddress)

		client := redis.NewClient(&redis.Options{
			Addr:            opt.Addr,
			MaxRetries:      3,
			MaxRetryBackoff: 1 * time.Minute,
			MinRetryBackoff: 5 * time.Second,
			// IdleTimeout: 5 * time.Minute,
			Password: "",     // no password set
			DB:       opt.DB, // use default DB
		})

		redisClient = &RedisClient{client}
		pong, _ := redisClient.Ping().Result()
		if pong != "" {
			log.Println("Connected to redis")
		}
	})

	return redisClient
}

func (c *RedisClient) set(key string, value string) {
	err := c.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}
func (c *RedisClient) get(key string) string {
	val, err := c.Get(key).Result()
	if err != nil {
		panic(err)
	}
	return val
}

// SetHost - Set Host
func (c *RedisClient) SetHost(value string) {
	c.set(host, value)
}

// GetHost - Get Host
func (c *RedisClient) GetHost() string {
	return c.get(host)
}

// SetPrevHost - Set Previes Host
func (c *RedisClient) SetPrevHost(value string) {
	c.set(prevHost, value)
}

// GetPrevHost - Get Previes Host
func (c *RedisClient) GetPrevHost() string {
	return c.get(prevHost)
}
