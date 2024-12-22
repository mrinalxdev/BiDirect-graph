package graphdb

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/mrinalxdev/bidirect/pkg/models"
)

type Partition struct {
	ID int
	RedisClient *redis.Client
}

func NewPartition(id int, redisAddr string) *Partition {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Partition{
		ID : id,
		RedisClient: client,
	}
}

func GetPartitionID(memberID models.MemberID, totalPartitions int) int {
	hash := sha256.Sum256([]byte(string(memberID)))
	return int(binary.BigEndian.Uint64(hash[:8]) % uint64(totalPartitions))
}

func (p *Partition) StoreConnection(ctx context.Context, conn models.Connection) error {
    // Store connections in Redis using sorted sets
    key := fmt.Sprintf("connections:%d", conn.SourceID)
    return p.RedisClient.ZAdd(ctx, key, &redis.Z{
        Score:  float64(conn.DestID),
        Member: conn.DestID,
    }).Err()
}

func (p *Partition) GetConnections(ctx context.Context, memberID models.MemberID) ([]models.MemberID, error) {
    key := fmt.Sprintf("connections:%d", memberID)
    results, err := p.RedisClient.ZRange(ctx, key, 0, -1).Result()
    if err != nil {
        return nil, err
    }

    connections := make([]models.MemberID, len(results))
    for i, result := range results {
        memberID, _ := strconv.ParseInt(result, 10, 64)
        connections[i] = models.MemberID(memberID)
    }
    return connections, nil
}