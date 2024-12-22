package ncs

type GraphNode struct {
    ID         string
    Partitions map[int]struct{}
}

func FindMinimumNodeSet(requiredPartitions map[int]struct{}, nodes []GraphNode) []GraphNode {
    var result []GraphNode
    uncovered := make(map[int]struct{})
    

    for partition := range requiredPartitions {
        uncovered[partition] = struct{}{}
    }

    for len(uncovered) > 0 {
        bestNode := findBestCoveringNode(uncovered, nodes)
        if bestNode == nil {
            break
        }
        
        result = append(result, *bestNode)
        for partition := range bestNode.Partitions {
            delete(uncovered, partition)
        }
    }
    
    return result
}

func findBestCoveringNode(uncovered map[int]struct{}, nodes []GraphNode) *GraphNode {
    var bestNode *GraphNode
    maxCovered := 0

    for i := range nodes {
        covered := 0
        for partition := range nodes[i].Partitions {
            if _, exists := uncovered[partition]; exists {
                covered++
            }
        }
        
        if covered > maxCovered {
            maxCovered = covered
            bestNode = &nodes[i]
        }
    }
    
    return bestNode
}