package token

import (
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/panjf2000/ants"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math"
	"math/big"
	"net/http"
	"sync"
)

type TokenPoolTask struct {
	TokenABI *eth.Token
	Token    *common.Token
	Owner    string
}

func OwnedListHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)

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
INNER JOIN ucoin.users AS u ON (u.wallet_addr = t.owner)
WHERE u.id = %d
ORDER BY t.updated_at DESC`, user.Id)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	var (
		wg     sync.WaitGroup
		tokens []*common.Token
	)
	pool, _ := ants.NewPoolWithFunc(10, func(val interface{}) error {
		defer wg.Done()
		task := val.(TokenPoolTask)
		tokenABI := task.TokenABI
		token := task.Token
		_, _, _, _, token.TotalSupply, token.TotalTransfers, token.TotalHolders, token.CirculatingSupply, token.Balance, err = utils.TokenMeta(tokenABI, task.Owner)
		return nil
	})

	for _, row := range rows {
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
		if token.TxStatus == 1 {
			tokenABI, err := utils.NewToken(token.Address, Service.Geth)
			if CheckErr(err, c) {
				raven.CaptureError(err, nil)
				return
			}
			task := TokenPoolTask{
				TokenABI: tokenABI,
				Token:    token,
				Owner:    user.Wallet,
			}
			wg.Add(1)
			pool.Serve(task)
		}
		tokens = append(tokens, token)
	}

	wg.Wait()

	c.JSON(http.StatusOK, tokens)
}
