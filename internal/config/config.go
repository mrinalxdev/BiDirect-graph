package config

import (
    "strings"
    "os"
)

type RedisConfig struct {
    Addresses []string
}

func LoadRedisConfig() RedisConfig {
    // Get Redis nodes from environment variable
    redisNodes := os.Getenv("REDIS_NODES")
    if redisNodes == "" {
        redisNodes = "localhost:6379"
    }

    return RedisConfig{
        Addresses: strings.Split(redisNodes, ","),
    }
}