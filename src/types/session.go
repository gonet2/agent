package types

import (
	"net"
	"time"
)

import (
	"misc/crypto/pike"
)

const (
	SESS_LOGGED_IN  = 0x1  // 玩家是否登录
	SESS_KICKED_OUT = 0x2  // 玩家是否被服务器踢掉
	SESS_REGISTERED = 0x4  // 是否已经执行过注册(避免恶意注册)
	SESS_KEYEXCG    = 0x8  // 是否已经交换完毕KEY
	SESS_ENCRYPT    = 0x10 // 是否可以开始加密
)

type Session struct {
	IP            net.IP
	MQ            chan IPCObject // 玩家消息队列(系统到玩家，玩家到玩家）
	Encoder       *pike.Pike     // 加密器
	Decoder       *pike.Pike     // 解密器
	ClientVersion int32          // 客户端版本

	// TODO : 玩家数据
	UserId int32
	User   *User //玩家数据

	// 会话标记
	Flag int32

	// 时间相关
	ConnectTime    time.Time // TCP链接建立时间
	PacketTime     time.Time // 当前包的到达时间
	LastPacketTime time.Time // 前一个包到达时间

	// RPS控制
	PacketCount uint32 // 对收到的包进行计数，避免恶意发包

	// ipo相关
	AppId        string
	OSVersion    string
	DeviceName   string
	DeviceId     string
	DeviceIdType int32

	_dirtycount int32
}

func (sess *Session) MarkDirty() {
	sess._dirtycount++
}

func (sess *Session) DirtyCount() int32 {
	return sess._dirtycount
}

func (sess *Session) MarkClean() {
	sess._dirtycount = 0
}
