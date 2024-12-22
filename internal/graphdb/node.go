package graphdb

import (
	"context"
	"fmt"
	"sync"

	// "github.com/go-redis/redis/v8"
	"github.com/mrinalxdev/bidirect/pkg/models"
)

type Node struct {
	ID string
	Partition map[int]*Partition
	mu sync.RWMutex
}

type NodeConfig struct {
	ID string
	PartitionIDs []int
	RedisAddr string
	ReplicaFactor int
}

func NewNode(config NodeConfig) *Node {
	node := &Node{
		ID : config.ID,
		Partition: make(map[int]*Partition),
	}

	for _, pid := range config.PartitionIDs{
		node.Partition[pid] = NewPartition(pid, config.RedisAddr)
	}

	return node
}

// retrives and merges second-degree connections for
// given number Id's from node's particition
func (n *Node) GetSecondDegreeConnections(ctx context.Context, memberIDs []models.MemberID) ([]models.MemberID, error) {
    n.mu.RLock()
    defer n.mu.RUnlock()
    uniqueConnections := make(map[models.MemberID]struct{})
    
    for _, memberID := range memberIDs {
        partitionID := GetPartitionID(memberID, len(n.Partition))
        partition, exists := n.Partition[partitionID]
        if !exists {
            continue
        }

        connections, err := partition.GetConnections(ctx, memberID)
        if err != nil {
            return nil, fmt.Errorf("failed to get connections for member %d: %w", memberID, err)
        }

        for _, conn := range connections {
            uniqueConnections[conn] = struct{}{}
        }
    }

    result := make([]models.MemberID, 0, len(uniqueConnections))
    for conn := range uniqueConnections {
        result = append(result, conn)
    }

    return result, nil
}