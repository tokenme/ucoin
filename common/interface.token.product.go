package common

import (
	"math/big"
)

type TokenProduct struct {
	Address      string   `json:"address"`
	Owner        string   `json:"owner,omitempty"`
	Title        string   `json:"title"`
	Symbol       string   `json:"symbol,omitempty"`
	TotalSupply  *big.Int `json:"total_supply,omitempty"`
	Token        *Token   `json:"token,omitempty"`
	Price        *big.Int `json:"price"`
	Amount       uint     `json:"amount"`
	TxStatus     int      `json:"tx_status"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Images       []string `json:"images,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Description  string   `json:"desc,omitempty"`
	OnlineStatus int      `json:"online_status,omitempty"`
}
