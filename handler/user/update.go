package user

import (
	//"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/utils"
	"net/http"
	"strings"
)

type UpdateRequest struct {
	Mobile      string         `form:"mobile" json:"mobile"`
	CountryCode uint           `form:"country_code" json:"country_code"`
	VerifyCode  string         `form:"verify_code" json:"verify_code"`
	Password    string         `form:"passwd" json:"passwd"`
	RePassword  string         `form:"repasswd" json:"repasswd"`
	Realname    string         `form:"realname" json:"realname"`
	Wechat      *WechatRequest `form:"wechat" json:"wechat"`
}

func UpdateHandler(c *gin.Context) {
	var req UpdateRequest
	if CheckErr(c.Bind(&req), c) {
		return
	}
	userContext, exists := c.Get("USER")
	if CheckWithCode(!exists, UNAUTHORIZED_ERROR, "need login", c) {
		return
	}
	user := userContext.(common.User)
	db := Service.Db

	var updateFields []string
	if req.Realname != "" {
		updateFields = append(updateFields, fmt.Sprintf("realname='%s'", db.Escape(req.Realname)))
	}
	var (
		unionId       string
		passwd        string
		mobile        string
		migrateUserId uint64
	)
	if req.Mobile != "" {
		if Check(req.Mobile == "" || req.CountryCode == 0 || req.VerifyCode == "" || req.Password == "" || req.RePassword == "", "missing params", c) {
			return
		}
		if Check(req.Password != req.RePassword, "repassword!=password", c) {
			return
		}
		passwdLength := len(req.Password)
		if Check(passwdLength < 8 || passwdLength > 64, "password length must between 8-32", c) {
			return
		}
		token, err := uuid.NewV4()
		if CheckErr(err, c) {
			return
		}
		salt := utils.Sha1(token.String())
		passwd = utils.Sha1(fmt.Sprintf("%s%s%s", salt, req.Password, salt))
		mobile = strings.Replace(req.Mobile, " ", "", 0)
		rows, _, err := db.Query(`SELECT 1 FROM ucoin.auth_verify_codes WHERE country_code=%d AND mobile='%s' AND code='%s' LIMIT 1`, req.CountryCode, db.Escape(mobile), db.Escape(req.VerifyCode))
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		if Check(len(rows) == 0, "unverified phone number", c) {
			return
		}
		updateFields = append(updateFields, fmt.Sprintf("country_code=%d, mobile='%s', salt='%s', passwd='%s'", req.CountryCode, db.Escape(mobile), db.Escape(salt), db.Escape(passwd)))
	} else if req.Wechat != nil {
		if Check(req.Wechat.AccessKey == "", "missing wx_session_key", c) {
			return
		}
		rows, _, err := db.Query(`SELECT union_id, session_key FROM ucoin.wx_oauth WHERE k='%s' LIMIT 1`, db.Escape(req.Wechat.AccessKey))
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
		if Check(len(rows) == 0, "unauthorized", c) {
			return
		}
		unionId = rows[0].Str(0)
		updateFields = append(updateFields, fmt.Sprintf("wx_unionid='%s', wx_nick='%s', wx_avatar='%s', wx_gender=%d, wx_city='%s', wx_province='%s', wx_country='%s', wx_language='%s'", db.Escape(unionId), db.Escape(req.Wechat.Nick), db.Escape(req.Wechat.Avatar), req.Wechat.Gender, db.Escape(req.Wechat.City), db.Escape(req.Wechat.Province), db.Escape(req.Wechat.Country), db.Escape(req.Wechat.Language)))
	}
	if mobile != "" && unionId != "" {
		if user.Mobile == "" && user.Wechat != nil && user.Wechat.UnionId != unionId {
			rows, _, err := db.Query(`SELECT id FROM ucoin.users WHERE wx_unionid='%s' LIMIT 1`, db.Escape(unionId))
			if CheckErr(err, c) {
				raven.CaptureError(err, nil)
				return
			}
			if Check(len(rows) == 0, "not found", c) {
				return
			}
			migrateUserId = rows[0].Uint64(0)
		} else if user.Wechat == nil && user.CountryCode != req.CountryCode && user.Mobile != mobile {
			rows, _, err := db.Query(`SELECT id FROM ucoin.users WHERE country_code='%d' AND mobile='%s' LIMIT 1`, req.CountryCode, db.Escape(mobile))
			if CheckErr(err, c) {
				raven.CaptureError(err, nil)
				return
			}
			if Check(len(rows) == 0, "not found", c) {
				return
			}
			migrateUserId = rows[0].Uint64(0)
		}
	}
	if len(updateFields) == 0 {
		c.JSON(http.StatusOK, APIResponse{Msg: "ok"})
		return
	}

	if migrateUserId > 0 {
		err := migrateUserAccount(user.Id, migrateUserId)
		if CheckErr(err, c) {
			raven.CaptureError(err, nil)
			return
		}
	}

	_, _, err := db.Query(`UPDATE ucoin.users SET %s WHERE id=%d LIMIT 1`, strings.Join(updateFields, ","), user.Id)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	c.JSON(http.StatusOK, APIResponse{Msg: "ok"})
}

func migrateUserAccount(toId uint64, fromId uint64) error {
	db := Service.Db
	_, _, err := db.Query(`DELETE FROM ucoin.users WHERE id=%d`, fromId)
	return err
}
