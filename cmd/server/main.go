package main

import (
	// "context"
	"fmt"
	"log"
	"net/http"
	// "os"
	// "os/signal"
	// "strings"
	// "syscall"
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

        nodeConfig := graphdb.NodeConfig{
            ID : fmt.Sprintf("node-%d", i),
            PartitionIDs: make([]int, 0, partitionsPerNode),
            RedisAddr: addr,
            ReplicaFactor: len(redisConfig.Addresses),
        }

        for pid := startPartition; pid < endPartitions; pid++{
            nodeConfig.PartitionIDs = append(nodeConfig.PartitionIDs, pid)
            logPartitionsInfo(nodeConfig.ID, pid, addr)
        }

        node := graphdb.NewNode(nodeConfig)
        if node == nil {
            log.Fatalf("Failed to create node %d", i)
        }

        for _, pid := range nodeConfig.PartitionIDs {
            if partition, exists := node.Partition[pid]; !exists || partition == nil {
                log.Fatalf("Faile to initialize partition %d in node %s", pid, node.ID)
            }
        }

        nodes[i] = node
    }

    log.Printf("Verifuing partition distribution ...")
    for i := 0; i < totalPartitions; i++ {
        found := false
        for _, node := range nodes {
            if _, exists := node.Partition[i]; exists {
                found = true
                log.Printf("Partition %D found in node %s", i, node.ID)
                break
            }
        }
        if !found {
            log.Printf("Warning : Partition %d is not assigned to any node", i)
        }
    }

    networkCache := ncs.NetworkCache(redisConfig.Addresses[0], 24*time.Hour)
    handler := api.NewHandler(nodes, networkCache)

    log.Printf("Starting server with %d nodes and %d partitions per node", 
        len(nodes), partitionsPerNode)
    log.Fatal(http.ListenAndServe(":8080", handler))
}