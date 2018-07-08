package auth

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/nu7hatch/gouuid"
	"github.com/tokenme/ucoin/common"
	. "github.com/tokenme/ucoin/handler"
	"github.com/tokenme/ucoin/utils"
	"io/ioutil"
	"net/http"
)

const WX_AUTH_GATEWAY = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

func WechatHandler(c *gin.Context) {
	code := c.Query("code")
	if Check(code == "", "missing wechat oauth code", c) {
		return
	}
	resp, err := http.Get(fmt.Sprintf(WX_AUTH_GATEWAY, Config.WXAppId, Config.WXSecret, code))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	var oauth common.WechatOAuth
	err = json.Unmarshal(body, &oauth)
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	if Check(oauth.UnionId == "" || oauth.SessionKey == "", "invalid auth info", c) {
		return
	}
	token, err := uuid.NewV4()
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	accessKey := utils.Sha1(fmt.Sprintf("%s%s%s", token, oauth.UnionId, oauth.SessionKey))
	db := Service.Db
	_, _, err = db.Query(`INSERT INTO ucoin.wx_oauth (union_id, k, session_key) VALUES ('%s', '%s', '%s') ON DUPLICATE KEY UPDATE k=VALUES(k), session_key=VALUES(session_key)`, db.Escape(oauth.UnionId), db.Escape(accessKey), db.Escape(oauth.SessionKey))
	if CheckErr(err, c) {
		raven.CaptureError(err, nil)
		return
	}
	oauth.AccessKey = accessKey
	c.JSON(http.StatusOK, oauth)
}
