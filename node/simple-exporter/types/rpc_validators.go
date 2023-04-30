package types

// RpcValidatorsResponse represents the response json structure of the /validators endpoint
type RpcValidatorsResponse struct {
	Result struct {
		BlockHeight string         `json:"block_height"`
		Validators  []RpcValidator `json:"validators"`
		Count       string         `json:"count"`
		Total       string         `json:"total"`
	} `json:"result"`
}

// RpcValidator represents the structure of a given RpcValidatorsResponse validator
type RpcValidator struct {
	Address          string   `json:"address"`
	VotingPower      string   `json:"voting_power"`
	ProposerPriority string   `json:"proposer_priority"`
	PubKey           struct { // Note: this represents the consensus pub_key
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
}
