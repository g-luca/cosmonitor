package main

import (
	"flag"
	"log"
	"os"
	"simple-exporter/core"
	"simple-exporter/prometheus"
)

var (
	// Define string, int, and bool flags
	nodeRpc = flag.String("node_rpc", "", "RPC endpoint of the wanted node (ex. https://rpc.cosmos.network:443)")
)

func main() {
	var rpcAddr = ""
	// get from the env
	if os.Getenv("NODE_RPC") != "" {
		rpcAddr = os.Getenv("NODE_RPC")
	} else {
		flag.Parse() // parse the command flags
		// ensure valid node_rpc endpoint flags
		if nodeRpc == nil || *nodeRpc == "" {
			log.Fatal("Not valid -node_rpc flag.")
		}
		rpcAddr = *nodeRpc
	}

	log.Printf("Running RPC node: %s", rpcAddr)

	go prometheus.StartPrometheus(9090)

	core.ListenWS(rpcAddr)
}
