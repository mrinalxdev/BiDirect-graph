package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mrinalxdev/bidirect/internal/api"
	"github.com/mrinalxdev/bidirect/internal/config"
	"github.com/mrinalxdev/bidirect/internal/graphdb"
	"github.com/mrinalxdev/bidirect/internal/ncs"
)

func main(){
    redisConfig := config.LoadRedisConfig()
    partitionsPerNode := 10
    totalPartitions := len(redisConfig.Addresses) * partitionsPerNode

    nodes := make([]*graphdb.Node, len(redisConfig.Addresses))
    logPartitionsInfo := func(nodeID string, partitionID int, addr string){
        log.Printf("Initializing - Noe : %s, Partitions: %d, Redis: %s", nodeID, partitionID, addr)
    }

    for i, addr := range redisConfig.Addresses{
        startPartition := i * partitionsPerNode
        endPartitions := startPartition + partitionsPerNode

        nodeCofig := graphdb.NodeConfig{
            ID : fmt.Sprintf("node-%d", i),
            PartitionIDs: make([]int, 0, partitionsPerNode),
            RedisAddr: addr,
            ReplicaFactor: len(redisConfig.Addresses),
        }

        for pid := startPartition; pid < endPartitions; pid++{
            
        }
    }
}