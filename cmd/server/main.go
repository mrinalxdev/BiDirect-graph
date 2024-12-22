package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mrinalxdev/bidirect/internal/graphdb"
	"github.com/mrinalxdev/bidirect/internal/ncs"
)

func main(){
		partitions := make([]*graphdb.Partition, 10)
		for i := range partitions {
			partitions[i] = graphdb.NewPartition(i, "localhost:6379")
		}
	
		networkCache := ncs.NewNetworkCache("localhost:6379", 24*time.Hour)
		handler := api.NewHandler(partitions, networkCache)
	
		log.Fatal(http.ListenAndServe(":8080", handler))
}