package abci

import (
	"context"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

// GetValidatorSigningInfo queries the ABCI endpoint to get the SigningInfo of a given Validator
func GetValidatorSigningInfo(rpc string, validatorAddr string) (*slashingTypes.ValidatorSigningInfo, error) {
	var client, _ = tmhttp.New(rpc, "")

	// prepare the request data
	var request = slashingTypes.QuerySigningInfoRequest{
		ConsAddress: validatorAddr,
	}
	data, _ := request.Marshal()

	var ctx = context.Background()
	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	// perform the ABCI query
	raw, err := ABCIQuery(bctx, client, "/cosmos.slashing.v1beta1.Query/SigningInfo", data)
	defer cancel()
	if err != nil || raw.Response.Log != "" {
		return nil, err
	}

	// decode the response
	var signingInfo slashingTypes.QuerySigningInfoResponse
	err = signingInfo.Unmarshal(raw.Response.GetValue())
	if err != nil {
		return nil, err
	}

	// return the wanted data
	return &signingInfo.ValSigningInfo, nil

}

// GetValidators queries the ABCI endpoint to get the Validators
func GetValidators(rpc string) (*[]stakingTypes.Validator, error) {
	var client, _ = tmhttp.New(rpc, "")
	var nextKey []byte
	var done = false

	var validators []stakingTypes.Validator

	for done == false {
		// prepare the request data and pagination
		var request = stakingTypes.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Key:        nextKey,
				Limit:      200,
				CountTotal: true,
				Reverse:    false,
			},
		}
		data, _ := request.Marshal()

		var ctx = context.Background()
		bctx, cancel := context.WithTimeout(ctx, 10*time.Second)

		// perform the ABCI query
		raw, err := ABCIQuery(bctx, client, "/cosmos.staking.v1beta1.Query/Validators", data)
		defer cancel()
		if err != nil || raw.Response.Log != "" {
			return nil, err
		}

		// decode the response
		var validatorsRes stakingTypes.QueryValidatorsResponse
		err = validatorsRes.Unmarshal(raw.Response.GetValue())
		if err != nil {
			return nil, err
		}

		// handle the new pagination
		if validatorsRes.Pagination.NextKey != nil {
			nextKey = validatorsRes.Pagination.NextKey
		} else {
			done = true
		}

		// extract only the wanted data
		for _, validator := range validatorsRes.Validators {
			validators = append(validators, validator)
		}
	}

	return &validators, nil

}

// GetValidatorCommission queries the ABCI endpoint to get the Validator commissions
func GetValidatorCommission(rpc string, validatorAddr string) (*types.DecCoins, error) {
	var client, _ = tmhttp.New(rpc, "")

	// prepare the request data
	var request = distributionTypes.QueryValidatorCommissionRequest{
		ValidatorAddress: validatorAddr,
	}
	data, _ := request.Marshal()

	var ctx = context.Background()
	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	// perform the ABCI query
	raw, err := ABCIQuery(bctx, client, "/cosmos.distribution.v1beta1.Query/ValidatorCommission", data)
	defer cancel()
	if err != nil || raw.Response.Log != "" {
		return nil, err
	}

	// decode the response
	var response distributionTypes.QueryValidatorCommissionResponse
	err = response.Unmarshal(raw.Response.GetValue())
	if err != nil {
		return nil, err
	}

	// return the wanted data
	return &response.Commission.Commission, nil

}

// GetValidatorRewards queries the ABCI endpoint to get the Validator rewards
func GetValidatorRewards(rpc string, validatorAddr string) (*types.DecCoins, error) {
	var client, _ = tmhttp.New(rpc, "")

	// prepare the request data
	var request = distributionTypes.QueryValidatorOutstandingRewardsRequest{
		ValidatorAddress: validatorAddr,
	}
	data, _ := request.Marshal()

	var ctx = context.Background()
	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	// perform the ABCI query
	raw, err := ABCIQuery(bctx, client, "/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards", data)
	defer cancel()
	if err != nil || raw.Response.Log != "" {
		return nil, err
	}

	// decode the response
	var response distributionTypes.QueryValidatorOutstandingRewardsResponse
	err = response.Unmarshal(raw.Response.GetValue())
	if err != nil {
		return nil, err
	}

	// return the wanted data
	return &response.Rewards.Rewards, nil

}
