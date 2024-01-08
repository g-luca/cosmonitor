package core

import (
	"errors"
	"fmt"
	bech322 "github.com/cosmos/cosmos-sdk/types/bech32"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/types"
	"log"
	"simple-exporter/abci"
	"simple-exporter/prometheus"
	"simple-exporter/rpc"
	"time"
)

func ListenWS(rpcAddr string) {
	// create the ABCI client
	client, err := tmhttp.New(rpcAddr, "")
	if err != nil {
		log.Println(err.Error())
	}

	defer func() {
		err = client.Stop()
		if err != nil {
			log.Println(err.Error())
		}
	}()

	var retry = 0
	const retryTimeout = 10
	for true {
		time.Sleep(3 * time.Second)
		var err = UpdateMetrics(client)
		if err != nil {
			log.Println(err.Error())
			prometheus.UpdateNodeInfo(false, "", "", "")
			retry += 1
			time.Sleep(retryTimeout * time.Second)
			log.Println(fmt.Sprintf("Error Updating metrics, retrying in %ds (attempt #%d)", retryTimeout, retry))
			continue
		} else {
			prometheus.DeleteNodeInfo("", "", "")
		}
		log.Println("Metrics updated correctly")
	}
}

func UpdateMetrics(client *tmhttp.HTTP) error {
	// get the node info
	nodeInfo, err := rpc.GetNodeInfo(client)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("Fetched node '%s' info", nodeInfo.NodeInfo.Moniker))
	log.Println(fmt.Sprintf("Network: '%s'", nodeInfo.NodeInfo.Network))

	// get the Validators from Consensus
	consValidators, err := rpc.GetValidators(client)
	if err != nil {
		return err
	}

	prometheus.UpdateNodeInfo(true, nodeInfo.NodeInfo.Network, nodeInfo.NodeInfo.Moniker, string(nodeInfo.NodeInfo.DefaultNodeID))

	// get the wanted Consensus validator
	var consValidator, rank = getConsValidatorFromAddress(nodeInfo.ValidatorInfo.Address.String(), consValidators)
	if consValidator == nil {
		log.Println(fmt.Sprintf("Node '%s' is not a Validator", nodeInfo.NodeInfo.Moniker))
		return nil
	}

	// set the generic Validators Consensus info (that don't need Validators ABCI Query)
	prometheus.UpdateTotalVotingPower(uint64(calculateTotalVotingPower(consValidators)))
	prometheus.UpdateVotingPower(uint64(consValidator.VotingPower))
	prometheus.UpdateRank(rank)

	// retrieve chain Bech32 Prefix from the ABCI endpoint (since v0.46)
	bech32Prefix, err := abci.GetBech32Prefix(client)
	if err != nil {
		// chain ot supported, try getting an address
		bech32Prefix, err = abci.GetBech32PrefixFromAuthAccounts(client)
		if err != nil {
			return err
		}
	}

	// calculate the validator "valcons" address
	valConsAddr, err := bech322.ConvertAndEncode(bech32Prefix+"valcons", consValidator.Address.Bytes())
	if err != nil {
		return err
	}

	// update signing info
	signingInfo, err := abci.GetValidatorSigningInfo(client, valConsAddr)
	if err != nil {
		return err
	}
	prometheus.UpdateMissedBlocks(signingInfo.MissedBlocksCounter)
	prometheus.UpdateTombstoned(signingInfo.Tombstoned)

	// get the Validator info from the ABCI Queries
	// NOTE: ICS Consumer chains may not have this endpoints
	abciValidators, err := abci.GetValidators(client)
	if err != nil {
		println("Cannot get ABCI Validator Info (ICS chain)")
		prometheus.UpdateValidatorInfo(true, nodeInfo.NodeInfo.Moniker, "", valConsAddr)
		return nil
	}

	// retrieve the validator from the abci validators query from the consensus one
	var wantedValidator = retrieveValidator(consValidator, abciValidators)
	if wantedValidator == nil {
		return errors.New("cannot retrieve Validator from ConsValidator")
	}

	// validator generic info
	prometheus.UpdateValidatorInfo(true, wantedValidator.GetMoniker(), wantedValidator.OperatorAddress, valConsAddr)

	// validator details
	prometheus.UpdateCommissionMaxChangeRate(wantedValidator.Commission.MaxChangeRate.MustFloat64())
	prometheus.UpdateCommissionMaxRate(wantedValidator.Commission.MaxRate.MustFloat64())
	prometheus.UpdateCommissionRate(wantedValidator.Commission.Rate.MustFloat64())
	prometheus.UpdateDelegatedTokens(wantedValidator.Tokens.Uint64())
	prometheus.UpdateJailed(wantedValidator.Jailed)
	prometheus.UpdateUnbondingHeight(wantedValidator.UnbondingHeight)

	// validator signing info
	prometheus.UpdateMinSelfDelegation(wantedValidator.MinSelfDelegation.Uint64())

	// get the validator commission
	commission, err := abci.GetValidatorCommission(client, wantedValidator.OperatorAddress)
	if err != nil {
		return err
	}
	prometheus.UpdateValidatorCommission(commission)

	// get the validator rewards
	rewards, err := abci.GetValidatorRewards(client, wantedValidator.OperatorAddress)
	if err != nil {
		return err
	}
	prometheus.UpdateValidatorRewards(rewards)

	return nil
}

func calculateTotalVotingPower(consValidators *[]*ctypes.Validator) int64 {
	var totalVotingPower int64 = 0
	for _, val := range *consValidators {
		totalVotingPower += val.VotingPower
	}
	return totalVotingPower
}
func getConsValidatorFromAddress(address string, consValidators *[]*ctypes.Validator) (*ctypes.Validator, int) {
	var validator *ctypes.Validator = nil

	for i, v := range *consValidators {
		if v.Address.String() == address {
			return v, i + 1
		}
	}
	return validator, -100
}

// retrieveValidator retrieves the current validator inside the Validators from the ConsValidator
func retrieveValidator(consValidator *ctypes.Validator, validators *[]stakingTypes.Validator) *stakingTypes.Validator {
	for _, validator := range *validators {
		var keyBytes = validator.ConsensusPubkey.Value
		if len(keyBytes) > 32 {
			keyBytes = keyBytes[2:]
		}

		vsPubKey := consValidator.PubKey.Bytes()
		if string(keyBytes) == string(vsPubKey) {
			return &validator
		}
	}
	return nil
}
