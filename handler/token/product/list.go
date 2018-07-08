package product

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
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
	"time"
)

const DefaultPageSize uint = 10

type ProductPoolTask struct {
	ABI     *eth.NFToken
	Product *common.TokenProduct
	Owner   string
}

type ListRequest struct {
	Token    string `form:"token" json:"token"`
	Page     uint   `form:"page" json:"page"`
	PageSize uint   `form:"page_size" json:"page_size"`
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
	if req.Token != "" {
		wheres = append(wheres, fmt.Sprintf("erc20.address='%s'", db.Escape(req.Token)))
	}

	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}
	rows, _, err := db.Query(`SELECT
	erc20.address, 
	erc20.owner, 
	erc20.name, 
	erc20.symbol, 
	erc20.decimals,
	erc20.total_supply AS initial_supply,
	erc20.logo,
	erc20.summary,
	t.address, 
	t.owner, 
	t.name, 
	t.price, 
	t.amount,
	IFNULL(txs.status, 2) AS tx_status,
	t.start_date,
	t.end_date,
	t.images,
	t.tags,
	t.summary,
	t.online_status
FROM ucoin.erc721 AS t 
INNER JOIN ucoin.users AS u ON (u.wallet_addr = t.owner)
INNER JOIN ucoin.erc20 AS erc20 ON (erc20.address = t.erc20)
LEFT JOIN ucoin.txs AS txs ON (txs.tx = t.tx)
%s
ORDER BY t.updated_at DESC LIMIT %d, %d`, where, page*pageSize, pageSize)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}

	var (
		wg       sync.WaitGroup
		products []*common.TokenProduct
	)

	pool, _ := ants.NewPoolWithFunc(10, func(val interface{}) error {
		defer wg.Done()
		task := val.(ProductPoolTask)
		tokenABI := task.ABI
		product := task.Product
		totalSupply, err := tokenABI.TotalSupply(nil)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		product.TotalSupply = totalSupply
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
			Logo:          row.Str(6),
			Description:   row.Str(7),
		}
		erc721 := &common.TokenProduct{
			Address:      row.Str(8),
			Owner:        row.Str(9),
			Title:        row.Str(10),
			Price:        new(big.Int).SetUint64(row.Uint64(11)),
			Amount:       row.Uint(12),
			TxStatus:     row.Uint(13),
			StartDate:    row.ForceLocaltime(14).Format(time.RFC3339),
			EndDate:      row.ForceLocaltime(15).Format(time.RFC3339),
			Description:  row.Str(18),
			Token:        token,
			OnlineStatus: row.Int(19),
		}
		imgArr := strings.Split(row.Str(16), ",")
		for _, img := range imgArr {
			if img != "" {
				erc721.Images = append(erc721.Images, img)
			}
		}
		tagArr := strings.Split(row.Str(17), " ")
		for _, tag := range tagArr {
			if tag != "" {
				erc721.Tags = append(erc721.Tags, tag)
			}
		}
		tokenABI, err := utils.NewNFToken(erc721.Address, Service.Geth)
		if CheckErr(err, c) {
			return
		}
		task := ProductPoolTask{
			ABI:     tokenABI,
			Product: erc721,
		}
		wg.Add(1)
		pool.Serve(task)

		products = append(products, erc721)
	}
	wg.Wait()
	c.JSON(http.StatusOK, products)
}
