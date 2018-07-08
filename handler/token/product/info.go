package product

import (
	//"github.com/davecgh/go-spew/spew"
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
	"time"
)

type InfoRequest struct {
	Address string `form:"address" json:"address" binding:"required"`
}

func InfoHandler(c *gin.Context) {
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
	t.address, 
	t.owner, 
	t.name, 
	t.price, 
	t.amount,
	t.start_date,
	t.end_date,
	t.images,
	t.tags,
	t.summary,
	t.online_status
FROM ucoin.erc721 AS t 
INNER JOIN ucoin.users AS u ON (u.wallet_addr = t.owner)
INNER JOIN ucoin.erc20 AS erc20 ON (erc20.address = t.erc20)
WHERE t.address = '%s'`, db.Escape(req.Address))
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
		Address:      row.Str(8),
		Owner:        row.Str(9),
		Title:        row.Str(10),
		Price:        new(big.Int).SetUint64(row.Uint64(11)),
		Amount:       row.Uint(12),
		StartDate:    row.ForceLocaltime(13).Format(time.RFC3339),
		EndDate:      row.ForceLocaltime(14).Format(time.RFC3339),
		Description:  row.Str(17),
		Token:        token,
		OnlineStatus: row.Int(18),
	}
	imgArr := strings.Split(row.Str(15), ",")
	for _, img := range imgArr {
		if img != "" {
			erc721.Images = append(erc721.Images, img)
		}
	}
	tagArr := strings.Split(row.Str(16), " ")
	for _, tag := range tagArr {
		if tag != "" {
			erc721.Tags = append(erc721.Tags, tag)
		}
	}
	tokenABI, err := utils.NewNFToken(erc721.Address, Service.Geth)
	if CheckErr(err, c) {
		return
	}
	totalSupply, err := tokenABI.TotalSupply(nil)
	if err != nil {
		log.Error(err.Error())
	}
	erc721.TotalSupply = totalSupply
	c.JSON(http.StatusOK, erc721)
}
