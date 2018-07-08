package order

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	//"github.com/mkideal/log"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const DefaultPageSize uint = 10

type ListRequest struct {
	Product   string `form:"product" json:"product"`
	OwnerType uint   `form:"owner_type" json:"owner_type"`
	Page      uint   `form:"page" json:"page"`
	PageSize  uint   `form:"page_size" json:"page_size"`
}

func ListHandler(c *gin.Context) {
	var req ListRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	var page uint
	if req.Page > 0 {
		page = req.Page
	}
	var pageSize = DefaultPageSize
	if req.PageSize > 0 && req.PageSize < DefaultPageSize {
		pageSize = req.PageSize
	}

	db := Service.Db
	var (
		where  string
		wheres []string
	)
	if req.Product != "" {
		wheres = append(wheres, fmt.Sprintf("o.erc721='%s'", db.Escape(req.Product)))
	}

	if req.OwnerType == 1 {
		userContext, exists := c.Get("USER")
		if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
			return
		}
		user := userContext.(common.User)
		wheres = append(wheres, fmt.Sprintf("o.buyer='%s'", db.Escape(user.Wallet)))
	} else if req.OwnerType == 2 {
		userContext, exists := c.Get("USER")
		if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
			return
		}
		user := userContext.(common.User)
		wheres = append(wheres, fmt.Sprintf("o.sellet='%s'", db.Escape(user.Wallet)))
	}

	if CheckWithCode(len(wheres) == 0, BADREQUEST_ERROR, "bad request", c) {
		return
	}
	where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
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
%s
ORDER BY o.id DESC LIMIT %d, %d`, where, page*pageSize, pageSize)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	var orders []common.Order
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

		if req.Product != "" {
			erc721.Token = nil
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
		orders = append(orders, order)
	}
	c.JSON(http.StatusOK, orders)
}
