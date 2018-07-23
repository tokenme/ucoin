package qrcode

import (
	//"github.com/davecgh/go-spew/spew"
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"math/big"
	"net/http"
)

type CollectRequest struct {
	Token  string `form:"token" json:"token"`
	Amount uint64 `form:"amount" json:"amount"`
}

func CollectHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req CollectRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	var token common.Token
	if req.Token != "" {
		db := Service.Db
		rows, _, err := db.Query(`SELECT address, name, symbol, decimals, total_supply AS initial_supply, logo FROM ucoin.erc20 WHERE address='%s' LIMIT 1`, db.Escape(req.Token))
		if CheckErr(err, c) {
			return
		}
		if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "token not found", c) {
			return
		}
		row := rows[0]
		token = common.Token{
			Address:     row.Str(0),
			Name:        row.Str(1),
			Symbol:      row.Str(2),
			Decimals:    row.Uint(3),
			TotalSupply: new(big.Int).SetUint64(row.Uint64(4)),
			Logo:        row.Str(5),
		}
	}
	code := common.NewCollectCode(user.Wallet, token, req.Amount)
	qrcode, err := code.String(Config.QRCodeUrl, []byte(Config.LinkSalt))
	if CheckErr(err, c) {
		return
	}
	c.JSON(http.StatusOK, APIResponse{Msg: qrcode})
}
