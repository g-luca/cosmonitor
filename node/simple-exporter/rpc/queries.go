package rpc

import (
	"context"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	ctypes "github.com/tendermint/tendermint/types"
)

// GetNodeInfo queries the RPC endpoint /status to get the node info
func GetNodeInfo(client *tmhttp.HTTP) (*coretypes.ResultStatus, error) {
	// perform the /status request
	resp, err := client.Status(context.Background())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetValidators queries the RPC endpoint /validators
func GetValidators(client *tmhttp.HTTP) (*[]*ctypes.Validator, error) {
	var requestedHeight int64 = 0
	var requestedPage = 1 // starts from 1
	var perPage = 100
	var totalEntries = 0
	var done = false

	var validators []*ctypes.Validator

	for !done {
		// perform the /validators request
		resp, err := client.Validators(context.Background(), nil, &requestedPage, &perPage)
		if err != nil {
			return nil, err
		}
		// append the validators
		for _, validator := range resp.Validators {
			validators = append(validators, validator)
		}

		// set the requested height only if first cycle
		if requestedHeight == 0 {
			requestedHeight = resp.BlockHeight
		}
		// set the total entries only if first cycle
		if totalEntries == 0 {
			totalEntries = resp.Total
		}

		// if the validators count matches with the total, it is done
		if len(validators) >= totalEntries {
			done = true
			break
		}

		// go to the next page
		requestedPage += 1
	}

	return &validators, nil
}
