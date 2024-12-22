package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrinalxdev/bidirect/internal/api"
	"github.com/mrinalxdev/bidirect/internal/config"
	"github.com/mrinalxdev/bidirect/internal/graphdb"
	"github.com/mrinalxdev/bidirect/internal/ncs"
)

func main() {
    // Load configuration
    redisConfig := config.LoadRedisConfig()

    // Initialize GraphDB nodes
    nodes := make([]*graphdb.Node, len(redisConfig.Addresses))
    partitionsPerNode := 10 // You can adjust this number

    for i, addr := range redisConfig.Addresses {
        // Calculate partition IDs for this node
        partitionIDs := make([]int, partitionsPerNode)
        for j := 0; j < partitionsPerNode; j++ {
            partitionIDs[j] = i*partitionsPerNode + j
        }

        // Create node configuration
        nodeConfig := graphdb.NodeConfig{
            ID:            fmt.Sprintf("node-%d", i),
            PartitionIDs:  partitionIDs,
            RedisAddr:     addr,
            ReplicaFactor: len(redisConfig.Addresses),
        }

        nodes[i] = graphdb.NewNode(nodeConfig)
    }

    networkCache := ncs.NewNetworkCache(redisConfig.Addresses[0], 24*time.Hour)

    // Initialize API handlers
    handler := api.NewHandler(nodes, networkCache)

    // Start HTTP server
    log.Printf("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}