package prometheus

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/prometheus/client_golang/prometheus"
)

// Define custom metrics for Node info
var (
	nodeInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "node_info",
			Help: "Node Info",
		},
		[]string{"network", "moniker", "id"},
	)
)

// Define custom metrics for Validator info
var (
	validatorInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "validator_info",
			Help: "Validator info",
		},
		[]string{"moniker", "valoper", "valcons"},
	)
)

// Define custom metrics from Consensus
var (
	totalVotingPower = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_voting_power",
		Help: "Total Validators Voting Power",
	})
	votingPower = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_voting_power",
		Help: "Validator Voting Power",
	})
)

// Define custom metrics from Info
var (
	jailed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_jailed",
		Help: "Validator Jailed Status",
	})
	rank = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_rank",
		Help: "Validator Rank",
	})
	minSelfDelegation = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_min_self_delegation",
		Help: "Validator Minimum Self Delegation",
	})
	delegatedTokens = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_delegated_tokens",
		Help: "Validator Delgated Tokens",
	})
	unbondingHeight = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_unbonding_height",
		Help: "Validator Unbonding Height",
	})
	commissionMaxChangeRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_commission_max_change_rate",
		Help: "Validator Max Change Rate",
	})
	commissionMaxRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_commission_max_rate",
		Help: "Validator Max Rate",
	})
	commissionRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_commission_rate",
		Help: "Validator Rate",
	})
)

// Define custom metrics from Signing Info
var (
	tombstoned = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_tombstoned",
		Help: "Validator Tombstoned",
	})
	missedBlocks = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "validator_missed_blocks",
		Help: "Validator Missed Blocks",
	})
)

// Define custom metric for balances
var (
	validatorCommission = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "validator_commission",
			Help: "Validator Commission",
		},
		[]string{"denom"},
	)
	validatorRewards = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "validator_rewards",
			Help: "Validator Rewards",
		},
		[]string{"denom"},
	)
)

func UpdateNodeInfo(isOnline bool, network string, moniker string, id string) {
	var onlineValue = 0
	if isOnline {
		onlineValue = 1
	}
	nodeInfo.WithLabelValues(network, moniker, id).Set(float64(onlineValue))
}

func DeleteNodeInfo(network string, moniker string, id string) {
	nodeInfo.DeleteLabelValues(network, moniker, id)
}

func UpdateValidatorInfo(isOnline bool, moniker string, valoper string, valcons string) {
	var onlineValue = 0
	if isOnline {
		onlineValue = 1
	}
	validatorInfo.WithLabelValues(moniker, valoper, valcons).Set(float64(onlineValue))
}

func UpdateRank(value int) {
	rank.Set(float64(value))
}

func UpdateTotalVotingPower(value uint64) {
	totalVotingPower.Set(float64(value))
}

func UpdateVotingPower(value uint64) {
	votingPower.Set(float64(value))
}

func UpdateJailed(isJailed bool) {
	if isJailed {
		jailed.Set(1)
	} else {
		jailed.Set(0)
	}
}

func UpdateMinSelfDelegation(value uint64) {
	minSelfDelegation.Set(float64(value))
}

func UpdateDelegatedTokens(value uint64) {
	delegatedTokens.Set(float64(value))
}

func UpdateUnbondingHeight(value int64) {
	unbondingHeight.Set(float64(value))
}

func UpdateTombstoned(isTombstoned bool) {
	if isTombstoned {
		tombstoned.Set(1)
	} else {
		tombstoned.Set(0)
	}
}

func UpdateMissedBlocks(value int64) {
	missedBlocks.Set(float64(value))
}

func UpdateCommissionRate(value float64) {
	commissionRate.Set(value)
}

func UpdateCommissionMaxRate(value float64) {
	commissionMaxRate.Set(value)
}

func UpdateCommissionMaxChangeRate(value float64) {
	commissionMaxChangeRate.Set(value)
}

func UpdateValidatorCommission(coins *types.DecCoins) {
	if coins == nil {
		return
	}
	for _, coin := range *coins {
		validatorCommission.WithLabelValues(coin.Denom).Set(coin.Amount.MustFloat64())
	}
}

func UpdateValidatorRewards(coins *types.DecCoins) {
	if coins == nil {
		return
	}
	for _, coin := range *coins {
		validatorRewards.WithLabelValues(coin.Denom).Set(coin.Amount.MustFloat64())
	}
}
