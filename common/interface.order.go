package common

import (
	"github.com/tokenme/ucoin/utils"
	"github.com/tokenme/ucoin/utils/binary"
	"math/big"
)

type Order struct {
	TokenId         uint64       `json:"token_id"`
	Buyer           string       `json:"buyer"`
	Seller          string       `json:"seller"`
	Product         TokenProduct `json:"product"`
	Price           *big.Int     `json:"price"`
	Tx              string       `json:"tx"`
	TokenTxStatus   uint         `json:"token_tx_status,omitempty"`
	ProductTxStatus uint         `json:"product_tx_status,omitempty"`
	InsertedAt      string       `json:"inserted_at"`
	UpdatedAt       string       `json:"udpated_at"`
	Qrcode          *Qrcode      `json:"qrcode,omitempty"`
}

type OrderProto struct {
	TokenId uint64 `json:"token_id"`
	Product string `json:"product"`
}

func (this Order) GetQrcode(key []byte) (*Qrcode, error) {
	proto := OrderProto{
		TokenId: this.TokenId,
		Product: this.Product.Address,
	}
	data, err := EncodeOrder(key, proto)
	if err != nil {
		return nil, err
	}
	return &Qrcode{
		Method: "/order/info",
		Data:   data,
	}, nil
}

func EncodeOrder(key []byte, proto OrderProto) (string, error) {
	enc := binary.NewEncoder()
	enc.Encode(proto)
	return utils.AESEncryptBytes(key, enc.Buffer())
}

func DecodeOrder(key []byte, cryptoText string) (proto OrderProto, err error) {
	data, err := utils.AESDecryptBytes(key, cryptoText)
	if err != nil {
		return proto, err
	}
	dec := binary.NewDecoder()
	dec.SetBuffer(data)
	dec.Decode(&proto)
	return proto, nil
}
