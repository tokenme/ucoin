package token

import (
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/mkideal/log"
	"github.com/tokenme/ucoin/coins/eth/utils"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math"
	"math/big"
	"net/http"
	"strings"
)

func InfoHandler(c *gin.Context) {
	tokenAddress := c.Param("address")
	if CheckWithCode(tokenAddress == "", NOTFOUND_ERROR, "NOT FOUND", c) {
		return
	}

	tokenAddress = strings.ToLower(tokenAddress)

	db := Service.Db
	rows, _, err := db.Query(`SELECT
	t.address, 
	t.owner, 
	t.name, 
	t.symbol, 
	t.decimals,
	t.total_supply AS initial_supply,
	IFNULL(txs.status, 2) AS tx_status,
	t.logo,
	t.summary
FROM ucoin.erc20 AS t 
LEFT JOIN ucoin.txs AS txs ON (txs.tx = t.tx)
WHERE t.address = '%s' LIMIT 1`, db.Escape(tokenAddress))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "NOT FOUND", c) {
		return
	}
	row := rows[0]
	var (
		decimals      = row.Uint(4)
		decimalsPow   = new(big.Int).SetUint64(uint64(math.Pow10(int(decimals))))
		initialSupply = new(big.Int).Mul(new(big.Int).SetUint64(row.Uint64(5)), decimalsPow)
	)
	token := &common.Token{
		Address:       row.Str(0),
		Owner:         row.Str(1),
		Name:          row.Str(2),
		Symbol:        row.Str(3),
		Decimals:      decimals,
		InitialSupply: initialSupply,
		TxStatus:      row.Uint(6),
		Logo:          row.Str(7),
		Description:   row.Str(8),
	}
	tokenABI, err := utils.NewToken(token.Address, Service.Geth)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	tokenMeta, err := token.GetMetaRedis(Service)
	if err == nil {
		token.TotalSupply = tokenMeta.TotalSupply
		token.TotalTransfers = tokenMeta.TotalTransfers
		token.CirculatingSupply = tokenMeta.CirculatingSupply
	} else {
		var foundError bool
		token.TotalSupply, err = tokenABI.TotalSupply(nil)
		if err != nil {
			log.Error(err.Error())
			foundError = true
		}
		token.TotalTransfers, err = tokenABI.TotalTransfers(nil)
		if err != nil {
			log.Error(err.Error())
			foundError = true
		}
		token.TotalHolders, err = tokenABI.TotalHolders(nil)
		if err != nil {
			log.Error(err.Error())
			foundError = true
		}
		token.CirculatingSupply, err = tokenABI.CirculatingSupply(nil)
		if err != nil {
			if !strings.Contains(err.Error(), "unmarshalling empty output") {
				log.Error(err.Error())
				foundError = true
			}
		}
		if !foundError {
			token.SaveMetaRedis(Service)
		}
	}
	userContext, exists := c.Get("USER")
	if exists {
		user := userContext.(common.User)
		token.Balance, err = utils.TokenBalanceOf(tokenABI, user.Wallet)
		if err != nil {
			log.Error(err.Error())
		}
	}
	c.JSON(http.StatusOK, token)
}
