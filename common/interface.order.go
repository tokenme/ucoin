package common

import (
	"github.com/tokenme/ucoin/utils"
	"github.com/tokenme/ucoin/utils/binary"
	"math/big"
)

type Order struct {
	TokenId         uint64       `json:"token_id"`
	Buyer           User         `json:"buyer"`
	Seller          User         `json:"seller"`
	Product         TokenProduct `json:"product"`
	Price           *big.Int     `json:"price"`
	Tx              string       `json:"tx"`
	TokenTxStatus   uint         `json:"token_tx_status,omitempty"`
	ProductTxStatus uint         `json:"product_tx_status,omitempty"`
	InsertedAt      string       `json:"inserted_at"`
	UpdatedAt       string       `json:"udpated_at"`
	Qrcode          string       `json:"qrcode,omitempty"`
}

type OrderProto struct {
	Qrcode
	TokenId uint64 `json:"token_id"`
	Product string `json:"product"`
}

func NewOrderProto(tokenId uint64, product string) OrderProto {
	return OrderProto{
		Qrcode: Qrcode{
			Method: QRCODE_ORDER_INFO,
		},
		TokenId: tokenId,
		Product: product,
	}
}

func (this OrderProto) String(host string, key []byte) (string, error) {
	data, err := this.Encode(key)
	if err != nil {
		return "", err
	}
	return this.Uri(host, data), nil
}

func (this OrderProto) Encode(key []byte) (string, error) {
	enc := binary.NewEncoder()
	enc.Encode(this)
	return utils.AESEncryptBytes(key, enc.Buffer())
}

func (this *OrderProto) Decode(key []byte, cryptoText string) error {
	data, err := utils.AESDecryptBytes(key, cryptoText)
	if err != nil {
		return err
	}
	dec := binary.NewDecoder()
	dec.SetBuffer(data)
	dec.Decode(this)
	return nil
}
