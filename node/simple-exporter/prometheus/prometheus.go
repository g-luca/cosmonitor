package prometheus

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func StartPrometheus(port uint) {
	// Register custom metrics with Prometheus
	prometheus.MustRegister(nodeInfo)
	prometheus.MustRegister(validatorInfo)
	prometheus.MustRegister(totalVotingPower)
	prometheus.MustRegister(rank)
	prometheus.MustRegister(votingPower)
	prometheus.MustRegister(jailed)
	prometheus.MustRegister(minSelfDelegation)
	prometheus.MustRegister(delegatedTokens)
	prometheus.MustRegister(unbondingHeight)
	prometheus.MustRegister(tombstoned)
	prometheus.MustRegister(missedBlocks)
	prometheus.MustRegister(commissionRate)
	prometheus.MustRegister(commissionMaxRate)
	prometheus.MustRegister(commissionMaxChangeRate)
	prometheus.MustRegister(validatorCommission)
	prometheus.MustRegister(validatorRewards)

	// Start an HTTP server to expose the metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Println(fmt.Sprintf("Starting Prometheus exporter on :%d/metrics", port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
