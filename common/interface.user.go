package common

import (
	"fmt"
	"github.com/tokenme/ucoin/utils"
)

type User struct {
	Id          uint64      `json:"id,omitempty"`
	CountryCode uint        `json:"country_code,omitempty"`
	Mobile      string      `json:"mobile,omitempty"`
	Nick        string      `json:"nick,omitempty"`
	Name        string      `json:"realname,omitempty"`
	Wechat      *WechatUser `json:"wechat,omitempty"`
	ShowName    string      `json:"showname,omitempty"`
	Avatar      string      `json:"avatar,omitempty"`
	Salt        string      `json:"-"`
	Password    string      `json:"-"`
	Wallet      string      `json:"wallet"`
	WalletPK    string      `json:"-"`
	CanPay      uint        `json:"can_pay,omitempty"`
}

type WechatUser struct {
	UnionId         string          `json:"unionId,omitempty"`
	OpenId          string          `json:"openId,omitempty"`
	Nick            string          `json:"nickName,omitempty"`
	Gender          uint            `json:"gender,omitempty"`
	City            string          `json:"city,omitempty"`
	Province        string          `json:"province,omitempty"`
	Country         string          `json:"country,omitempty"`
	Avatar          string          `json:"avatarUrl,omitempty"`
	Language        string          `json:"language,omitempty"`
	PhoneNumber     string          `json:"phoneNumber,omitempty"`
	PurePhoneNumber string          `json:"purePhoneNumber,omitempty"`
	CountryCode     string          `json:"countryCode,omitempty"`
	Watermark       WechatWatermark `json:"watermark,omitempty"`
}

type WechatWatermark struct {
	AppId     string `json:"appid,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type WechatOAuth struct {
	UnionId    string `json:"unionid,omitempty"`
	OpenId     string `json:"openid,omitempty"`
	SessionKey string `json:"session_key,omitempty"`
	AccessKey  string `json:"access_key, omitempty"`
}

func (this User) GetShowName() string {
	if this.Name != "" {
		return this.Name
	}
	if this.Nick != "" {
		return this.Nick
	}
	if this.Wechat != nil && this.Wechat.Nick != "" {
		return this.Wechat.Nick
	}
	return fmt.Sprintf("+%d%s", this.CountryCode, this.Mobile)
}

func (this User) GetAvatar(cdn string) string {
	if this.Avatar != "" {
		return this.Avatar
	}
	if this.Wechat != nil && this.Wechat.Avatar != "" {
		return this.Wechat.Avatar
	}
	key := utils.Md5(fmt.Sprintf("+%d%s", this.CountryCode, this.Mobile))
	return fmt.Sprintf("%suser/avatar/%s", cdn, key)
}
