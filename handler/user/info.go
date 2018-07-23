package user

import (
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"net/http"
)

func InfoGetHandler(c *gin.Context) {
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	if c.Query("refresh") != "" {
		db := Service.Db
		query := `SELECT 
                u.id, 
                u.country_code,
                u.mobile,
                u.nickname,
                u.avatar,
                u.realname,
                u.salt, 
                u.passwd,
                u.wallet_addr,
                u.wx_unionid,
                u.wx_nick,
                u.wx_avatar,
                u.payment_passwd
            FROM ucoin.users AS u
            LEFT JOIN ucoin.wx_oauth AS wx ON (wx.union_id=u.wx_unionid)
            WHERE u.id = %d
            AND active = 1
            LIMIT 1`
		rows, _, err := db.Query(query, user.Id)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		if CheckWithCode(len(rows) == 0, NOTFOUND_ERROR, "not found", c) {
			return
		}
		row := rows[0]
		user := common.User{
			Id:          row.Uint64(0),
			CountryCode: row.Uint(1),
			Mobile:      row.Str(2),
			Nick:        row.Str(3),
			Avatar:      row.Str(4),
			Name:        row.Str(5),
			Salt:        row.Str(6),
			Password:    row.Str(7),
			Wallet:      row.Str(8),
		}
		wxUnionId := row.Str(9)
		if wxUnionId != "" {
			user.Wechat = &common.WechatUser{
				UnionId: wxUnionId,
				Nick:    row.Str(10),
				Avatar:  row.Str(11),
			}
		}
		paymentPasswd := row.Str(12)
		if paymentPasswd != "" {
			user.CanPay = 1
		}
		user.ShowName = user.GetShowName()
		user.Avatar = user.GetAvatar(Config.CDNUrl)
	}
	c.JSON(http.StatusOK, user)
}
