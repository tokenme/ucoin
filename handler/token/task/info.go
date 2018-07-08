package task

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
	"strings"
	"time"
)

type InfoRequest struct {
	Id uint64 `form:"id" json:"id" binding:"required"`
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
    t.id, 
    t.owner, 
    t.name, 
    t.bonus, 
    t.amount,
    t.start_date,
    t.end_date,
    t.images,
    t.tags,
    t.summary,
    t.online_status,
    t.need_evidence
FROM ucoin.tasks AS t 
INNER JOIN ucoin.users AS u ON (u.wallet_addr = t.owner)
INNER JOIN ucoin.erc20 AS erc20 ON (erc20.address = t.erc20)
WHERE t.id = %d
LIMIT 1`, req.Id)
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
	task := common.TokenTask{
		Id:           row.Uint64(8),
		Owner:        row.Str(9),
		Title:        row.Str(10),
		Bonus:        new(big.Int).SetUint64(row.Uint64(11)),
		Amount:       row.Uint(12),
		StartDate:    row.ForceLocaltime(13).Format(time.RFC3339),
		EndDate:      row.ForceLocaltime(14).Format(time.RFC3339),
		Description:  row.Str(17),
		Token:        token,
		OnlineStatus: row.Int(18),
		NeedEvidence: row.Int(19),
	}
	imgArr := strings.Split(row.Str(15), ",")
	for _, img := range imgArr {
		if img != "" {
			task.Images = append(task.Images, img)
		}
	}
	tagArr := strings.Split(row.Str(16), " ")
	for _, tag := range tagArr {
		if tag != "" {
			task.Tags = append(task.Tags, tag)
		}
	}

	c.JSON(http.StatusOK, task)
}
