package task

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
	Token    string `form:"token" json:"token"`
	Page     uint   `form:"page" json:"page"`
	PageSize uint   `form:"page_size" json:"page_size"`
}

func ListHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	var user common.User

	if exists {
		user = userContext.(common.User)
	}

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
	wheres = append(wheres, fmt.Sprintf("(erc20.tx_status=1 AND t.online_status=1 OR t.owner='%s')", db.Escape(user.Wallet)))
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
%s
ORDER BY t.updated_at DESC LIMIT %d, %d`, where, page*pageSize, pageSize)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	var tasks []common.TokenTask
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
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, tasks)
}
