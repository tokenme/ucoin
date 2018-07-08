package qiniu

import (
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	"github.com/tokenme/ucoin/coins/eth"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/tools/qiniu"
	commonutils "github.com/tokenme/ucoin/utils"
	"net/http"
	"strconv"
	"time"
)

type TokenLogoRequest struct {
	Token string `form:"token" json:"token"`
}

func TokenLogoHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if Check(!exists, "need login", c) {
		return
	}
	user := userContext.(common.User)
	var req TokenLogoRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}

	if req.Token != "" {
		db := Service.Db
		rows, _, err := db.Query(`SELECT wallet, wallet_salt FROM ucoin.users WHERE id=%d LIMIT 1`, user.Id)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		if Check(len(rows) == 0, "user not exists", c) {
			return
		}
		row := rows[0]
		walletEncrypted := row.Str(0)
		walletSalt := row.Str(1)

		privKey, err := commonutils.AddressDecrypt(walletEncrypted, walletSalt, Config.TokenSalt)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}

		pubKey, err := eth.AddressFromHexPrivateKey(privKey)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}

		rows, _, err = db.Query(`SELECT address, name, symbol, decimals, logo FROM ucoin.erc20 WHERE address="%s" AND owner="%s" LIMIT 1`, db.Escape(req.Token), db.Escape(pubKey))
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		if Check(len(rows) == 0, "not found", c) {
			return
		}
	} else {
		tokenAddress, err := uuid.NewV4()
		if CheckErr(err, c) {
			return
		}
		req.Token = commonutils.Sha1(tokenAddress.String())
	}

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	upToken, key, link := qiniu.UpToken(Config.Qiniu, fmt.Sprintf("%s/%s", Config.Qiniu.LogoPath, req.Token), timestamp)
	c.JSON(http.StatusOK, gin.H{"uptoken": upToken, "key": key, "link": link})
}
