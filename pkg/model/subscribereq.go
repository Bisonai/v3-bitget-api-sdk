package model

type SubscribeReq struct {
	InstType string  `json:"instType,omitempty"`
	Channel  string  `json:"channel,omitempty"`
	InstId   *string `json:"instId,omitempty"`
	Coin     *string `json:"coin,omitempty"`
}
