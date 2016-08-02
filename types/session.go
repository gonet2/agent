package types

import (
	"crypto/rc4"
	"net"
	"time"
)

import (
	pb "agent/pb"
)

const (
	SESS_KEYEXCG    = 0x1 // 是否已经交换完毕KEY
	SESS_ENCRYPT    = 0x2 // 是否可以开始加密
	SESS_KICKED_OUT = 0x4 // 踢掉
	SESS_AUTHORIZED = 0x8 // 已授权访问
)

type Session struct {
	IP      net.IP
	MQ      chan pb.Game_Frame          // 返回给客户端的异步消息
	Encoder *rc4.Cipher                 // 加密器
	Decoder *rc4.Cipher                 // 解密器
	UserId  int32                       // 玩家ID
	GSID    string                      // 游戏服ID;e.g.: game1,game2
	Stream  pb.GameService_StreamClient // 后端游戏服数据流
	Die     chan struct{}               // 会话关闭信号

	// 会话标记
	Flag int32

	// 时间相关
	ConnectTime    time.Time // TCP链接建立时间
	PacketTime     time.Time // 当前包的到达时间
	LastPacketTime time.Time // 前一个包到达时间

	// RPS控制
	PacketCount uint32 // 对收到的包进行计数，避免恶意发包
}
