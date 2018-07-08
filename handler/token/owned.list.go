package token

import (
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/mkideal/log"
	"github.com/panjf2000/ants"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/coins/eth/utils"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math"
	"math/big"
	"net/http"
	"strings"
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
	IFNULL(txs.status, 2) AS tx_status,
	t.logo,
	t.summary
FROM ucoin.erc20 AS t 
INNER JOIN ucoin.users AS u ON (u.wallet_addr = t.owner)
LEFT JOIN ucoin.txs AS txs ON (txs.tx = t.tx)
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

		token.Balance, err = utils.TokenBalanceOf(tokenABI, task.Owner)
		if err != nil {
			log.Error(err.Error())
		}
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
			TxStatus:      row.Uint(6),
			Logo:          row.Str(7),
			Description:   row.Str(8),
		}

		tokenABI, err := utils.NewToken(token.Address, Service.Geth)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		task := TokenPoolTask{
			TokenABI: tokenABI,
			Token:    token,
			Owner:    token.Owner,
		}
		wg.Add(1)
		pool.Serve(task)
		tokens = append(tokens, token)
	}

	wg.Wait()

	c.JSON(http.StatusOK, tokens)
}
