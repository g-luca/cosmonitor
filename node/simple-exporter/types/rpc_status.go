package types

// RpcStatusResponse represents the response json structure of the /status endpoint
type RpcStatusResponse struct {
	Result struct {
		NodeInfo struct {
			ID         string `json:"id"`
			Network    string `json:"network"`
			Version    string `json:"version"`
			ListenAddr string `json:"listen_addr"`
			Moniker    string `json:"moniker"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash     string `json:"latest_block_hash"`
			LatestAppHash       string `json:"latest_app_hash"`
			LatestBlockHeight   string `json:"latest_block_height"`
			LatestBlockTime     string `json:"latest_block_time"`
			EarliestBlockHash   string `json:"earliest_block_hash"`
			EarliestAppHash     string `json:"earliest_app_hash"`
			EarliestBlockHeight string `json:"earliest_block_height"`
			EarliestBlockTime   string `json:"earliest_block_time"`
			CatchingUp          bool   `json:"catching_up"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address     string `json:"address"`
			VotingPower string `json:"voting_power"`
			PubKey      struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
		} `json:"validator_info"`
	} `json:"result"`
}
