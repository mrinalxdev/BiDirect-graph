package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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

// handles the distance calculation between members
func (h *Handler) GetDistances(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    var request struct {
        SourceID     models.MemberID   `json:"sourceId"`
        DestinationIDs []models.MemberID `json:"destinationIds"`
    }
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    secondDegree, err := h.networkCache.GetSecondDegree(request.SourceID)
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
    firstDegree, err := h.nodes[partitionID].GetPartitions()[partitionID].GetConnections(memberID)
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
            Partitions: node.Partitions,
        }
    }

    // Find minimum set of nodes needed
    selectedNodes := ncs.FindMinimumNodeSet(requiredPartitions, nodesList)
    var secondDegree []models.MemberID
    for _, node := range selectedNodes {
        connections, err := h.nodes[node.ID].GetSecondDegreeConnections(ctx, firstDegree)
        if err != nil {
            continue
        }
        secondDegree = append(secondDegree, connections...)
    }

    err = h.networkCache.StoreSecondDegree(memberID, secondDegree)
    if err != nil {
        // Log error but continue since we still have the results
        log.Printf("Failed to cache second degree connections: %v", err)
    }

    return secondDegree, nil
}

func (h *Handler) calculateDistance(ctx context.Context, source, dest models.MemberID, 
    secondDegree []models.MemberID) int {
    firstDegree, err := h.nodes[graphdb.GetPartitionID(source, len(h.nodes))].
        GetPartitions()[graphdb.GetPartitionID(source, len(h.nodes))].
        GetConnections(source)
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

    destConnections, err := h.nodes[graphdb.GetPartitionID(dest, len(h.nodes))].
        GetPartitions()[graphdb.GetPartitionID(dest, len(h.nodes))].
        GetConnections(dest)
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