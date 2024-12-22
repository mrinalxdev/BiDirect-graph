package ncs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mrinalxdev/bidirect/pkg/models"
)

type NetworkCache struct {
	redisClient *redis.Client
	ttl time.Duration
}

func NewNetworkCache(redisAddr string, ttl time.Duration) *NetworkCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &NetworkCache{
		redisClient: client,
		ttl: ttl,
	}
}

func (nc *NetworkCache) StoreSecondDegree(ctx context.Context, memberID models.MemberID, connections []models.MemberID) error {
	key := fmt.Sprintf("second_degree:%d", memberID)

	pipe := nc.redisClient.Pipeline()
	for _, conn := range connections{
		pipe.ZAdd(ctx, key, &redis.Z{
			Score: float64(conn),
			Member: conn,
		})
	}
	pipe.Expire(ctx, key, nc.ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (nc *NetworkCache) GetSecondDegree(ctx context.Context, memberId models.MemberID) ([] models.MemberID, error){
	key := fmt.Sprintf("second_degree:%d", memberId)
	results, err := nc.redisClient.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	connections := make([]models.MemberID, len(results))
	for i, result := range results {
		memberId, _ := strconv.ParseInt(result, 10, 64)
		connections[i] = models.MemberID(memberId)
	}

	return connections, nil
}