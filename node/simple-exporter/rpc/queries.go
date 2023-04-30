package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"simple-exporter/types"
	"strconv"
)

// GetNodeInfo queries the RPC endpoint /status to get the node info
func GetNodeInfo(rpc string) (*types.RpcStatusResponse, error) {
	resp, err := http.Get(rpc + "/status")
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	var status types.RpcStatusResponse
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		return nil, err
	}

	return &status, nil
}

// GetValidators queries the RPC endpoint /validators
func GetValidators(rpc string) (*[]types.RpcValidator, error) {
	var requestedHeight = 0
	var requestedPage = 1 // starts from 1
	var totalEntries = 0
	var done = false

	var validators []types.RpcValidator

	for done == false {
		var endpointUrl = fmt.Sprintf("%s/validators?per_page=100", rpc)
		// if a requested height is set append to the request endpoint url
		if requestedHeight > 0 {
			endpointUrl += fmt.Sprintf("&height=%d", requestedHeight)
		}
		// if a requested height is set append to the request endpoint url
		if requestedPage > 0 {
			endpointUrl += fmt.Sprintf("&page=%d", requestedPage)
		}
		// perform the request

		resp, err := http.Get(endpointUrl)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
		}

		var validatorsResponse types.RpcValidatorsResponse
		err = json.NewDecoder(resp.Body).Decode(&validatorsResponse)
		if err != nil {
			return nil, err
		}

		// append the validators
		for _, validator := range validatorsResponse.Result.Validators {
			validators = append(validators, validator)
		}

		// set the requested height only if first cycle
		if requestedHeight == 0 {
			requestedHeight, _ = strconv.Atoi(validatorsResponse.Result.BlockHeight)
		}
		// set the total entries only if first cycle
		if totalEntries == 0 {
			totalEntries, _ = strconv.Atoi(validatorsResponse.Result.Total)
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
