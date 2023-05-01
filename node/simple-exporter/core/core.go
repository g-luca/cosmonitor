package core

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/cosmos/btcutil/bech32"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	"log"
	"simple-exporter/abci"
	"simple-exporter/prometheus"
	"simple-exporter/rpc"
	"simple-exporter/types"
	"strconv"
	"strings"
	"time"
)

func _ListenWS(rpcAddr string) {

	var ctx = context.Background()

	// create the WebSocket client
	client, err := tmhttp.NewWithTimeout(rpcAddr, "/websocket", 5)
	if err != nil {
		log.Fatal(err.Error())
	}

	// expect a valid status query response
	_, err = client.Status(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	//with "https://rpc-cosmoshub-ia.cosmosia.notional.ventures:443" does work, why?
	// start the client
	err = client.Start()
	if err != nil {
		log.Fatal(err.Error())
	}

	// subscribe to the NewBlockHeader event (is lighter the NewBlock)
	subscribe, err := client.Subscribe(ctx, "exporter_subscribe", "tm.event = 'NewBlockHeader'")
	if err != nil {
		log.Fatal(err.Error())
	}

	// handle the new blocks
	for {
		select {
		case <-subscribe:
			log.Println("new block")
		}
	}

}

func ListenWS(rpcAddr string) {
	var retry = 0
	const retryTimeout = 10
	for true {
		//TODO: replace with receive block from WS -> UpdateMetrics()
		time.Sleep(3 * time.Second)
		var err = UpdateMetrics(rpcAddr)
		if err != nil {
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

func UpdateMetrics(rpcAddr string) error {
	// get the node info
	nodeInfo, err := rpc.GetNodeInfo(rpcAddr)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("Fetched node '%s' info", nodeInfo.Result.NodeInfo.Moniker))
	log.Println(fmt.Sprintf("Network: '%s'", nodeInfo.Result.NodeInfo.Network))

	// get the Validators from Consensus
	consValidators, err := rpc.GetValidators(rpcAddr)
	if err != nil {
		return err
	}

	prometheus.UpdateNodeInfo(true, nodeInfo.Result.NodeInfo.Network, nodeInfo.Result.NodeInfo.Moniker, nodeInfo.Result.NodeInfo.ID)

	// get the wanted Consensus validator
	var consValidator, rank = getConsValidatorFromAddress(nodeInfo.Result.ValidatorInfo.Address, consValidators)
	if consValidator == nil {
		log.Println(fmt.Sprintf("Node '%s' is not a Validator", nodeInfo.Result.NodeInfo.Moniker))
		return nil
	}

	// get the Validators from the Cosmos
	abciValidators, err := abci.GetValidators(rpcAddr)
	if err != nil {
		return err
	}

	// retrieve the validator from the abci validators query from the consensus one
	var wantedValidator = retrieveValidator(*consValidator, abciValidators)
	if wantedValidator == nil {
		return errors.New("cannot retrieve Validator from ConsValidator")
	}

	// compute the valcons address
	valConsAddr, err := computeValConsAddr(consValidator.Address, wantedValidator.OperatorAddress)
	if err != nil {
		return err
	}

	signingInfo, err := abci.GetValidatorSigningInfo(rpcAddr, *valConsAddr)
	if err != nil {
		return err
	}

	// validator generic info
	prometheus.UpdateValidatorInfo(true, wantedValidator.GetMoniker(), wantedValidator.OperatorAddress, *valConsAddr)

	// consensus info
	prometheus.UpdateTotalVotingPower(uint64(calculateTotalVotingPower(consValidators)))
	vp, err := strconv.Atoi(consValidator.VotingPower)
	if err != nil {
		vp = 0
	}
	prometheus.UpdateVotingPower(uint64(vp))

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
	commission, err := abci.GetValidatorCommission(rpcAddr, wantedValidator.OperatorAddress)
	if err != nil {
		return err
	}
	prometheus.UpdateValidatorCommission(commission)

	// get the validator rewards
	rewards, err := abci.GetValidatorRewards(rpcAddr, wantedValidator.OperatorAddress)
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

func calculateTotalVotingPower(consValidators *[]types.RpcValidator) int64 {
	var totalVotingPower int64 = 0
	for _, val := range *consValidators {
		vp, convErr := strconv.ParseInt(val.VotingPower, 10, 64)
		if convErr != nil {
			continue
		}
		totalVotingPower += vp
	}
	return totalVotingPower
}
func getConsValidatorFromAddress(address string, consValidators *[]types.RpcValidator) (*types.RpcValidator, int) {
	var validator *types.RpcValidator = nil

	for i, v := range *consValidators {
		if v.Address == address {
			return &v, i + 1
		}
	}
	return validator, -100
}

// calculateB64PubKeyFromVal returns the Base64 Public Key from the Validator struct
// Note: this is needed since `TmConsPublicKey()` uses the cachedValue which is not available from the ABCI query
func calculateB64PubKeyFromVal(validator *stakingTypes.Validator) string {
	var keyBytes = validator.ConsensusPubkey.Value
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[2:]
	}
	return base64.StdEncoding.EncodeToString(keyBytes)
}

// retrieveValidator retrieves the current validator inside the Validators from the ConsValidator
func retrieveValidator(consValidator types.RpcValidator, validators *[]stakingTypes.Validator) *stakingTypes.Validator {
	for _, validator := range *validators {
		pubKey := calculateB64PubKeyFromVal(&validator)
		if pubKey == consValidator.PubKey.Value {
			return &validator
		}
	}
	return nil
}
