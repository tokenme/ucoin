package common

import (
	"math/big"
)

type TokenTask struct {
	Id           uint64   `json:"id"`
	Owner        string   `json:"owner,omitempty"`
	Title        string   `json:"title"`
	Token        *Token   `json:"token,omitempty"`
	Bonus        *big.Int `json:"bonus"`
	Amount       uint     `json:"amount"`
	StartDate    string   `json:"start_date"`
	EndDate      string   `json:"end_date"`
	Images       []string `json:"images,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Description  string   `json:"desc,omitempty"`
	OnlineStatus int      `json:"online_status,omitempty"`
	NeedEvidence int      `json:"need_evidence,omitempty"`
}
