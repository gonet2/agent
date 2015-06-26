package types

const (
	AUTH_TYPE_UDID = iota
	AUTH_TYPE_WECHAT
	AUTH_TYPE_FB
)

type Auth struct {
	Id         int32  // user id
	Name       string // user name
	UniqueId   int64  // unique id use in chat, who will use chat like user;alliance;room etc
	Udid       string // client open udid
	Cert       string // binding account cert
	AuthType   int8   // auth type
	CreateTime int64  // account create time
}
