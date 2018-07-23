package common

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/big"
)

type Token struct {
	Address           string   `json:"address"`
	Owner             string   `json:"owner,omitempty"`
	Name              string   `json:"name"`
	Symbol            string   `json:"symbol"`
	Decimals          uint     `json:"decimals"`
	InitialSupply     *big.Int `json:"initial_supply,omitempty"`
	TotalSupply       *big.Int `json:"total_supply,omitempty"`
	TotalTransfers    *big.Int `json:"total_transfers,omitempty"`
	TotalHolders      *big.Int `json:"total_holders,omitempty"`
	CirculatingSupply *big.Int `json:"circulating_supply,omitempty"`
	Balance           *big.Int `json:"balance,omitempty"`
	TxStatus          int      `json:"tx_status"`
	Logo              string   `json:"logo,omitempty"`
	Cover             string   `json:"cover,omitempty"`
	Description       string   `json:"desc,omitempty"`
}

func (this Token) MetaRedisKey() string {
	return fmt.Sprintf("%s-meta", this.Address)
}

func (this Token) SaveMetaRedis(service *Service) error {
	redisMasterConn := service.Redis.Master.Get()
	defer redisMasterConn.Close()
	js, err := json.Marshal(this)
	if err != nil {
		return err
	}
	_, err = redisMasterConn.Do("SETEX", this.MetaRedisKey(), 60*60, js)
	return err
}

func (this Token) GetMetaRedis(service *Service) (Token, error) {
	redisMasterConn := service.Redis.Master.Get()
	defer redisMasterConn.Close()
	var ret Token
	buf, err := redis.Bytes(redisMasterConn.Do("GET", this.MetaRedisKey()))
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(buf, &ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (this Token) DeleteMetaRedis(service *Service) error {
	redisMasterConn := service.Redis.Master.Get()
	defer redisMasterConn.Close()
	_, err := redisMasterConn.Do("DEL", this.MetaRedisKey())
	return err
}
