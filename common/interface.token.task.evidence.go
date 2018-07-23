package common

import (
	"math/big"
)

type TokenTaskEvidence struct {
	Id            uint64    `json:"id"`
	Task          TokenTask `json:"task"`
	User          User      `json:"user"`
	Bonus         *big.Int  `json:"bonus"`
	Description   string    `json:"desc,omitempty"`
	Images        []string  `json:"images,omitempty"`
	CreateTime    string    `json:"create_time,omitempty"`
	UpdateTime    string    `json:"update_time,omitempty"`
	ApproveStatus int       `json:"approve_status,omitempty"`
	Tx            string    `json:"tx,omitempty"`
	TxStatus      uint      `json:"tx_status,omitempty"`
}
