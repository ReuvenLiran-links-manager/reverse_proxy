package main

import (
	"log"

	"github.com/go-redis/redis"
)

const host = "host"
const prevHost = "prev_host"
const fallbackAddress = "localhost:6379"

// Redis struct
type Redis struct {
	client redis.Client
}

// ConnectToRedis - connect to redis
func (r *Redis) ConnectToRedis() {
	client := *redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDRESS", fallbackAddress),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	r.client = client

	pong, err := r.client.Ping().Result()
	log.Printf(pong, err)
}

func (r *Redis) set(key string, value string) {
	err := r.client.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}
func (r *Redis) get(key string) string {
	val, err := r.client.Get(key).Result()
	if err != nil {
		panic(err)
	}
	return val
}

// SetHost - Set Host
func (r *Redis) SetHost(value string) {
	r.set(host, value)
}

// GetHost - Get Host
func (r *Redis) GetHost() string {
	return r.get(host)
}

// SetPrevHost - Set Previes Host
func (r *Redis) SetPrevHost(value string) {
	r.set(prevHost, value)
}

// GetPrevHost - Get Previes Host
func (r *Redis) GetPrevHost() string {
	return r.get(prevHost)
}
