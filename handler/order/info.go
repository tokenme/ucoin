package order

import (
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math"
	"math/big"
	"net/http"
	"time"
)

type InfoRequest struct {
	Id      uint64 `form:"id" json:"id"`
	Product string `form:"product" json:"product"`
}

func InfoHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req InfoRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}

	db := Service.Db
	rows, _, err := db.Query(`SELECT
    erc20.address, 
    erc20.owner, 
    erc20.name, 
    erc20.symbol, 
    erc20.decimals,
    erc20.total_supply AS initial_supply,
    erc20.logo,
    erc20.summary,
    erc721.address, 
    erc721.owner, 
    erc721.name, 
    erc721.start_date,
    erc721.end_date,
    o.buyer,
    o.seller,
    o.token_id,
    o.price,
    o.erc721_tx,
    o.inserted_at,
    o.updated_at,
    IFNULL(txs.status, 2) AS tx_status
FROM ucoin.erc721_orders AS o
INNER JOIN ucoin.erc721 AS erc721 ON (erc721.address = o.erc721)
INNER JOIN ucoin.erc20 AS erc20 ON (erc20.address = o.erc20)
LEFT JOIN ucoin.txs AS txs ON (txs.tx = o.erc721_tx)
WHERE o.erc721='%s' AND o.token_id=%d AND o.buyer='%s' OR o.seller='%s'
LIMIT 1`, db.Escape(req.Product), req.Id, db.Escape(user.Wallet), db.Escape(user.Wallet))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "not found", c) {
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
		Logo:          row.Str(6),
		Description:   row.Str(7),
	}
	erc721 := common.TokenProduct{
		Address:   row.Str(8),
		Owner:     row.Str(9),
		Title:     row.Str(10),
		StartDate: row.ForceLocaltime(11).Format(time.RFC3339),
		EndDate:   row.ForceLocaltime(12).Format(time.RFC3339),
		Token:     token,
	}
	order := common.Order{
		TokenId:         row.Uint64(15),
		Buyer:           row.Str(13),
		Seller:          row.Str(14),
		Price:           new(big.Int).SetUint64(row.Uint64(16)),
		Tx:              row.Str(17),
		InsertedAt:      row.ForceLocaltime(18).Format(time.RFC3339),
		UpdatedAt:       row.ForceLocaltime(19).Format(time.RFC3339),
		ProductTxStatus: row.Uint(20),
		Product:         erc721,
	}
	qrcode, err := order.GetQrcode([]byte(Config.LinkSalt))
	if CheckErr(err, c) {
		return
	}
	order.Qrcode = qrcode
	c.JSON(http.StatusOK, order)
}
