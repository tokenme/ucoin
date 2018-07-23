package common

type Config struct {
	AppName             string       `default:"ucoin"`
	BaseUrl             string       `default:"https://ucoin.io"`
	CDNUrl              string       `default:"https://cdn.ucoin.io/"`
	QRCodeUrl           string       `default:"qr.ucoin.io"`
	Port                int          `default:"8008"`
	Geth                string       `default:"geth.xibao100.com"`
	StaticPath          string       `required:"true"`
	LogPath             string       `required:"true"`
	TokenSalt           string       `required:"true"`
	LinkSalt            string       `required:"true"`
	OutputWallet        OutputWallet `required:"true"`
	ERC20Template       string       `required:"true"`
	ERC721Template      string       `required:"true"`
	ERC721Symbol        string       `required:"true"`
	SentryDSN           string       `required:"true"`
	MySQL               MySQLConfig  `required:"true"`
	Redis               RedisConfig  `required:"true"`
	SQS                 SQSConfig    `required:"true"`
	S3                  S3Config     `required:"true"`
	Qiniu               QiniuConfig  `required:"true"`
	TwilioToken         string       `required:"true"`
	WXAppId             string       `required:"true"`
	WXSecret            string       `required:"true"`
	SlackToken          string       `required:"true"`
	SlackAdminChannelID string       `required:"true"`
	GeoIP               string       `required:"true"`
	Debug               bool         `default:"false"`
	EnableWeb           bool         `default:"false"`
	EnableGC            bool         `default:"false"`
	EnableTx            bool         `default:"false"`
}

type MySQLConfig struct {
	Host   string `required:"true"`
	User   string `required:"true"`
	Passwd string `required:"true"`
	DB     string `default:"tokenme"`
}

type RedisConfig struct {
	Master string `required:"true"`
	Slave  string
}

type OutputWallet struct {
	Salt string `required:"true"`
	Data string `required:"true"`
}

type SQSConfig struct {
	Region    string `default:"ap-northeast-1"`
	AccountId string `required:"true"`
	AK        string `required:"true"`
	Secret    string `required:"true"`
	TxQueue   string `default:"ucoin-tx"`
	GasQueue  string `default:"ucoin-gas"`
	Token     string `default:""`
}

type S3Config struct {
	Region   string `default:"ap-northeast-1"`
	AK       string `required:"true"`
	Secret   string `required:"true"`
	Bucket   string `required:"true"`
	LogoPath string `requird:"true"`
	Token    string `default:""`
}

type QiniuConfig struct {
	AK                    string `required:"true"`
	Secret                string `required:"true"`
	Bucket                string `required:"true"`
	AvatarPath            string `required:"true"`
	LogoPath              string `required:"true"`
	TokenCover            string `required:"true"`
	TokenProductImagePath string `required:"true"`
	TokenTaskImagePath    string `required:"true"`
	TokenTaskEvidencePath string `required:"true"`
	Domain                string `required:"true"`
}
