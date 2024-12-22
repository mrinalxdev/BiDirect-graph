package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mrinalxdev/bidirect/internal/graphdb"
	"github.com/mrinalxdev/bidirect/internal/ncs"
	"github.com/mrinalxdev/bidirect/pkg/models"
)

type Handler struct {
    nodes        []*graphdb.Node
    networkCache *ncs.NetworkCache
    router       *mux.Router
}

func NewHandler(nodes []*graphdb.Node, networkCache *ncs.NetworkCache) *Handler {
    h := &Handler{
        nodes:        nodes,
        networkCache: networkCache,
    }

    router := mux.NewRouter()
    router.HandleFunc("/api/connections/{memberID}", h.GetConnections).Methods("GET")
    router.HandleFunc("/api/shared-connections/{memberID1}/{memberID2}", h.GetSharedConnections).Methods("GET")
    router.HandleFunc("/api/distances", h.GetDistances).Methods("POST")

    h.router = router
    return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.router.ServeHTTP(w, r)
}

func parseMemberID(idStr string) (models.MemberID, error) {
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid member ID format: %v", err)
    }
    return models.MemberID(id), nil
}


func (h *Handler) GetConnections(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    memberIDStr, ok := vars["memberID"]
    if !ok {
        http.Error(w, "memberID is required", http.StatusBadRequest)
        return
    }

    memberID, err := parseMemberID(memberIDStr)
    if err != nil {
        http.Error(w, fmt.Sprintf("Invalid member ID: %v", err), http.StatusBadRequest)
        return
    }
    
    // validate nodes array
    if len(h.nodes) == 0 {
        http.Error(w, "No nodes available", http.StatusInternalServerError)
        return
    }
    
    // Get partition ID with bounds checking
    partitionID := graphdb.GetPartitionID(memberID, len(h.nodes))
    if partitionID < 0 || partitionID >= len(h.nodes) {
        http.Error(w, "Invalid partition ID calculated", http.StatusInternalServerError)
        return
    }
    
    // Validate node exists
    if h.nodes[partitionID] == nil {
        http.Error(w, fmt.Sprintf("Node %d not initialized", partitionID), http.StatusInternalServerError)
        return
    }
    

    partition, exists := h.nodes[partitionID].Partition[partitionID]
    if !exists || partition == nil {
        http.Error(w, fmt.Sprintf("Partition %d not found", partitionID), http.StatusInternalServerError)
        return
    }
    

    connections, err := partition.GetConnections(ctx, memberID)
    if err != nil {
        log.Printf("Error getting connections for member %d: %v", memberID, err)
        http.Error(w, "Failed to retrieve connections", http.StatusInternalServerError)
        return
    }
    

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(connections); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}


func (h *Handler) GetSharedConnections(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    vars := mux.Vars(r)
    
    member1ID, err := parseMemberID(vars["memberID1"])
    if err != nil {
        http.Error(w, fmt.Sprintf("Invalid member ID 1: %v", err), http.StatusBadRequest)
        return
    }
    
    member2ID, err := parseMemberID(vars["memberID2"])
    if err != nil {
        http.Error(w, fmt.Sprintf("Invalid member ID 2: %v", err), http.StatusBadRequest)
        return
    }
    
    partition1 := graphdb.GetPartitionID(member1ID, len(h.nodes))
    partition2 := graphdb.GetPartitionID(member2ID, len(h.nodes))
    
    connections1, err := h.nodes[partition1].Partition[partition1].GetConnections(ctx, member1ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    connections2, err := h.nodes[partition2].Partition[partition2].GetConnections(ctx, member2ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Find shared connections
    shared := make([]models.MemberID, 0)
    connectionMap := make(map[models.MemberID]bool)
    
    for _, conn := range connections1 {
        connectionMap[conn] = true
    }
    
    for _, conn := range connections2 {
        if connectionMap[conn] {
            shared = append(shared, conn)
        }
    }
    
    json.NewEncoder(w).Encode(shared)
}

// handles the distance calculation between members
func (h *Handler) GetDistances(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    var request struct {
        SourceID       models.MemberID   `json:"sourceId"`
        DestinationIDs []models.MemberID `json:"destinationIds"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    secondDegree, err := h.networkCache.GetSecondDegree(ctx, request.SourceID)  // Updated with context
    if err != nil {
        secondDegree, err = h.buildSecondDegreeConnections(ctx, request.SourceID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    distances := make([]models.GraphDistance, 0, len(request.DestinationIDs))
    for _, destID := range request.DestinationIDs {
        distance := h.calculateDistance(ctx, request.SourceID, destID, secondDegree)
        distances = append(distances, models.GraphDistance{
            SourceID: request.SourceID,
            DestID:   destID,
            Distance: distance,
        })
    }

    json.NewEncoder(w).Encode(distances)
}

func (h *Handler) buildSecondDegreeConnections(ctx context.Context, memberID models.MemberID) ([]models.MemberID, error) {
    partitionID := graphdb.GetPartitionID(memberID, len(h.nodes))
    firstDegree, err := h.nodes[partitionID].Partition[partitionID].GetConnections(ctx, memberID)
    if err != nil {
        return nil, err
    }

    requiredPartitions := make(map[int]struct{})
    for _, conn := range firstDegree {
        pid := graphdb.GetPartitionID(conn, len(h.nodes))
        requiredPartitions[pid] = struct{}{}
    }

    nodesList := make([]ncs.GraphNode, len(h.nodes))
    for i, node := range h.nodes {
        nodesList[i] = ncs.GraphNode{
            ID:         node.ID,
            Partitions: make(map[int]struct{}),
        }
        for pid := range node.Partition {
            nodesList[i].Partitions[pid] = struct{}{}
        }
    }

    selectedNodes := ncs.FindMinimumNodeSet(requiredPartitions, nodesList)
    var secondDegree []models.MemberID
    
    for _, nodeInfo := range selectedNodes {
        for i, node := range h.nodes {
            if node.ID == nodeInfo.ID {
                connections, err := h.nodes[i].GetSecondDegreeConnections(ctx, firstDegree)
                if err != nil {
                    continue
                }
                secondDegree = append(secondDegree, connections...)
                break
            }
        }
    }

    err = h.networkCache.StoreSecondDegree(ctx, memberID, secondDegree)
    if err != nil {
        log.Printf("Failed to cache second degree connections: %v", err)
    }

    return secondDegree, nil
}

func (h *Handler) calculateDistance(ctx context.Context, source, dest models.MemberID, secondDegree []models.MemberID) int {
    partitionID := graphdb.GetPartitionID(source, len(h.nodes))
    firstDegree, err := h.nodes[partitionID].Partition[partitionID].GetConnections(ctx, source)
    if err == nil {
        for _, conn := range firstDegree {
            if conn == dest {
                return 1
            }
        }
    }

    for _, conn := range secondDegree {
        if conn == dest {
            return 2
        }
    }

    destPartitionID := graphdb.GetPartitionID(dest, len(h.nodes))
    destConnections, err := h.nodes[destPartitionID].Partition[destPartitionID].GetConnections(ctx, dest)
    if err == nil {
        for _, conn := range destConnections {
            for _, secondConn := range secondDegree {
                if conn == secondConn {
                    return 3
                }
            }
        }
    }

    return 4 // More than 3 degrees away
}