package abci

import (
	"context"
	"errors"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	"simple-exporter/types"
)

// ABCIQuery Perform an ABCI query
func ABCIQuery(ctx context.Context, client *tmhttp.HTTP, path string, data types.HexBytes) (*types.ResultABCIQuery, error) {
	if client == nil {
		return nil, errors.New("RPC Client not available")
	}
	response, err := client.ABCIQuery(ctx, path, []byte(data))
	if err != nil {
		return nil, err
	}
	return &types.ResultABCIQuery{
		Response: types.ResponseQuery{
			Code:      response.Response.Code,
			Log:       response.Response.Log,
			Info:      response.Response.Info,
			Index:     response.Response.Index,
			Key:       response.Response.Key,
			Value:     response.Response.Value,
			Height:    response.Response.Height,
			Codespace: response.Response.Codespace,
		},
	}, err
}
