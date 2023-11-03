package core

import (
	"errors"
	"fmt"
	"github.com/cosmos/btcutil/bech32"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/types"
	"log"
	"simple-exporter/abci"
	"simple-exporter/prometheus"
	"simple-exporter/rpc"
	"strings"
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

	// get the Validators from the Cosmos
	abciValidators, err := abci.GetValidators(client)
	if err != nil {
		return err
	}

	// retrieve the validator from the abci validators query from the consensus one
	var wantedValidator = retrieveValidator(consValidator, abciValidators)
	if wantedValidator == nil {
		return errors.New("cannot retrieve Validator from ConsValidator")
	}

	// compute the valcons address
	valConsAddr, err := computeValConsAddr(consValidator.Address.String(), wantedValidator.OperatorAddress)
	if err != nil {
		return err
	}

	signingInfo, err := abci.GetValidatorSigningInfo(client, *valConsAddr)
	if err != nil {
		return err
	}

	// validator generic info
	prometheus.UpdateValidatorInfo(true, wantedValidator.GetMoniker(), wantedValidator.OperatorAddress, *valConsAddr)

	// consensus info
	prometheus.UpdateTotalVotingPower(uint64(calculateTotalVotingPower(consValidators)))
	prometheus.UpdateVotingPower(uint64(consValidator.VotingPower))

	// validator info
	prometheus.UpdateRank(rank)
	prometheus.UpdateCommissionMaxChangeRate(wantedValidator.Commission.MaxChangeRate.MustFloat64())
	prometheus.UpdateCommissionMaxRate(wantedValidator.Commission.MaxRate.MustFloat64())
	prometheus.UpdateCommissionRate(wantedValidator.Commission.Rate.MustFloat64())
	prometheus.UpdateDelegatedTokens(wantedValidator.Tokens.Uint64())
	prometheus.UpdateJailed(wantedValidator.Jailed)
	prometheus.UpdateUnbondingHeight(wantedValidator.UnbondingHeight)

	// validator signing info
	prometheus.UpdateMinSelfDelegation(wantedValidator.MinSelfDelegation.Uint64())
	prometheus.UpdateMissedBlocks(signingInfo.MissedBlocksCounter)
	prometheus.UpdateTombstoned(signingInfo.Tombstoned)

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

// computeValConsAddr computes the valcons address from the consensus address and the operator address.
// The `operatorAddress` is needed to recover the HRP, the `consAddress` is needed to retrieve the address data bytes.
func computeValConsAddr(consAddress string, operatorAddress string) (*string, error) {
	// get the hrp
	hrp, _, err := bech32.Decode(operatorAddress, bech32.MaxLengthBIP173)
	if err != nil {
		return nil, err
	}
	// generate the valcons HRP from the valoper one
	var valConsHrp = strings.ReplaceAll(hrp, "valoper", "valcons")

	// Decode the consensus address hex string, this will generate a cosmsovalcons address
	consensusAddressBytes, err := sdkTypes.ConsAddressFromHex(consAddress)
	if err != nil {
		return nil, err
	}

	// get the data from it
	_, data, err := bech32.Decode(consensusAddressBytes.String(), bech32.MaxLengthBIP173)
	if err != nil {
		return nil, err
	}

	// regenerate the valcons Bech32 address with the wanted HRP
	valConsAddr, err := bech32.Encode(valConsHrp, data)
	if err != nil {
		return nil, err
	}

	return &valConsAddr, nil
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
