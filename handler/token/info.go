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
	userContext, exists := c.Get("USER")
	var user common.User
	if exists {
		user = userContext.(common.User)
	}
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
	t.tx_status,
	t.logo,
	t.summary
FROM ucoin.erc20 AS t 
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
		TxStatus:      row.Int(6),
		Logo:          row.Str(7),
		Description:   row.Str(8),
	}
	tokenABI, err := utils.NewToken(token.Address, Service.Geth)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	_, _, _, _, token.TotalSupply, token.TotalTransfers, token.TotalHolders, token.CirculatingSupply, token.Balance, err = utils.TokenMeta(tokenABI, user.Wallet)
	if err != nil {
		log.Error(err.Error())
	} else {
		token.SaveMetaRedis(Service)
	}
	c.JSON(http.StatusOK, token)
}
