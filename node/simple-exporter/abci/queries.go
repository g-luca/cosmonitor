package abci

import (
	"context"
	"errors"
	"fmt"
	"github.com/cosmos/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/rpc/client/http"
	"time"
)

// GetValidatorSigningInfo queries the ABCI endpoint to get the SigningInfo of a given Validator
func GetValidatorSigningInfo(client *http.HTTP, validatorAddr string) (*slashingTypes.ValidatorSigningInfo, error) {

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
		if raw.Response.Log != "" {
			return nil, errors.New(fmt.Sprintf("Invalid Response Log: %s", raw.Response.Log))
		}
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
func GetValidators(client *http.HTTP) (*[]stakingTypes.Validator, error) {
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
			if raw.Response.Log != "" {
				return nil, errors.New(fmt.Sprintf("Invalid Response Log: %s", raw.Response.Log))
			}
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
func GetValidatorCommission(client *http.HTTP, validatorAddr string) (*types.DecCoins, error) {
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
		if raw.Response.Log != "" {
			return nil, errors.New(fmt.Sprintf("Invalid Response Log: %s", raw.Response.Log))
		}
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
func GetValidatorRewards(client *http.HTTP, validatorAddr string) (*types.DecCoins, error) {

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
		if raw.Response.Log != "" {
			return nil, errors.New(fmt.Sprintf("Invalid Response Log: %s", raw.Response.Log))
		}
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

// GetBech32Prefix queries the endpoint to get the Bech32 prefix used for addresses generation
// Note: Available since Cosmos-Sdk v0.46
func GetBech32Prefix(client *http.HTTP) (string, error) {

	// prepare the request data
	var request = authTypes.Bech32PrefixRequest{}
	data, _ := request.Marshal()

	var ctx = context.Background()
	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	// perform the ABCI query
	raw, err := ABCIQuery(bctx, client, "/cosmos.auth.v1beta1.Query/Bech32Prefix", data)
	defer cancel()
	if err != nil || raw.Response.Log != "" {
		if raw.Response.Log != "" {
			return "", errors.New(fmt.Sprintf("Invalid Response Log: %s", raw.Response.Log))
		}
		return "", err
	}

	// decode the response
	var response authTypes.Bech32PrefixResponse
	err = response.Unmarshal(raw.Response.GetValue())
	if err != nil {
		return "", err
	}

	// return the wanted data
	return response.Bech32Prefix, nil

}

// GetBech32PrefixFromAuthAccounts queries the ABCI Bank accounts endpoint to get the first available Auth account
func GetBech32PrefixFromAuthAccounts(client *http.HTTP) (string, error) {

	// prepare the request data
	var request = authTypes.QueryAccountsRequest{
		Pagination: &query.PageRequest{
			Limit:   1,
			Reverse: true,
		},
	}
	data, _ := request.Marshal()

	var ctx = context.Background()
	bctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	// perform the ABCI query
	raw, err := ABCIQuery(bctx, client, "/cosmos.auth.v1beta1.Query/Accounts", data)
	defer cancel()
	if err != nil || raw.Response.Log != "" {
		if raw.Response.Log != "" {
			return "", errors.New(fmt.Sprintf("Invalid Response Log: %s", raw.Response.Log))
		}
		return "", err
	}

	// decode the response
	var response authTypes.QueryAccountResponse
	err = response.Unmarshal(raw.Response.GetValue())
	if err != nil {
		return "", err
	}
	switch response.Account.TypeUrl {
	case "/cosmos.auth.v1beta1.BaseAccount":
		var acc authTypes.BaseAccount
		err = acc.Unmarshal(response.Account.Value)
		hrp, _, err := bech32.Decode(acc.Address, bech32.MaxLengthBIP173)
		if err != nil {
			return "", err
		}
		return hrp, nil
	}
	// return the wanted data
	return "", errors.New(fmt.Sprintf("Cannot get Bech32 Account from account type %s", response.Account.TypeUrl))

}
