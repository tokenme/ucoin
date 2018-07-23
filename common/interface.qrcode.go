package common

import (
	"errors"
	"github.com/tokenme/ucoin/utils"
	"github.com/tokenme/ucoin/utils/binary"
	"net/url"
	"strings"
)

const (
	QRCODE_COLLECT    = "collect"
	QRCODE_ORDER_INFO = "order/info"
)

type QrcodeData interface {
	Encode(key []byte) (string, error)
	Decode(key []byte, cryptoText string) error
	String(host string, key []byte) (string, error)
}

type Qrcode struct {
	Method string `json:"method,omitempty"`
}

func (this Qrcode) Uri(host string, data string) string {
	uri := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   this.Method,
	}
	q := url.Values{}
	q.Set("q", data)
	uri.RawQuery = q.Encode()
	return uri.String()
}

func ParseQrcode(key []byte, uri string) (ret QrcodeData, error error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	method := strings.TrimPrefix(parsed.Path, "/")
	data := parsed.Query().Get("q")
	switch method {
	case QRCODE_COLLECT:
		ret = &CollectCode{}
	case QRCODE_ORDER_INFO:
		ret = &OrderProto{}
	default:
		return nil, errors.New("invalid code")
	}
	ret.Decode(key, data)
	return ret, nil
}

type CollectCode struct {
	Qrcode
	Wallet string `json:"wallet"`
	Token  Token  `json:"token,omitempty"`
	Amount uint64 `json:"amount,omitempty"`
}

func NewCollectCode(wallet string, token Token, amount uint64) CollectCode {
	return CollectCode{
		Qrcode: Qrcode{
			Method: QRCODE_COLLECT,
		},
		Wallet: wallet,
		Token:  token,
		Amount: amount,
	}
}

func (this CollectCode) String(host string, key []byte) (string, error) {
	data, err := this.Encode(key)
	if err != nil {
		return "", err
	}
	return this.Uri(host, data), nil
}

func (this CollectCode) Encode(key []byte) (string, error) {
	enc := binary.NewEncoder()
	enc.Encode(this)
	return utils.AESEncryptBytes(key, enc.Buffer())
}

func (this *CollectCode) Decode(key []byte, cryptoText string) error {
	data, err := utils.AESDecryptBytes(key, cryptoText)
	if err != nil {
		return err
	}
	dec := binary.NewDecoder()
	dec.SetBuffer(data)
	dec.Decode(this)
	return nil
}
