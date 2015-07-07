package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/peterbourgon/g2s"
)

import (
	"client_handler"
	"misc/packet"
	. "types"
	"utils"
)

const (
	STATSD_PREFIX       = "API.NR"
	ENV_STATSD          = "STATSD_HOST"
	DEFAULT_STATSD_HOST = "127.0.0.1:8125"
)

var (
	_statter g2s.Statter
)

func init() {
	addr := DEFAULT_STATSD_HOST
	if env := os.Getenv(ENV_STATSD); env != "" {
		addr = env
	}

	s, err := g2s.Dial("udp", addr)
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
	_statter = s
}

// client protocol handle proxy
func proxy_user_request(sess *Session, p []byte) {
	start := time.Now()
	defer utils.PrintPanicStack(sess, p)
	// 解密
	if sess.Flag&SESS_ENCRYPT != 0 {
		sess.Decoder.XORKeyStream(p, p)
	}

	// 封装为reader
	reader := packet.Reader(p)

	// 读客户端数据包序列号(1,2,3...)
	// 可避免重放攻击-REPLAY-ATTACK
	seq_id, err := reader.ReadU32()
	if err != nil {
		log.Error("read client timestamp failed:", err)
		sess.Flag |= SESS_KICKED_OUT
		return
	}

	// 读协议号
	b, err := reader.ReadS16()
	if err != nil {
		log.Error("read protocol number failed.")
		sess.Flag |= SESS_KICKED_OUT
		return
	}

	// 数据包序列号验证
	if seq_id != sess.PacketCount {
		log.Errorf("illegal packet sequence id:%v should be:%v proto:%v size:%v", seq_id, sess.PacketCount, b, len(p)-6)
		sess.Flag |= SESS_KICKED_OUT
		return
	}

	var ret []byte
	if b > MAX_PROTO_NUM { // game协议
		// 透传
		err = forward(sess, p)
		if err != nil {
			log.Errorf("service id:%v execute failed", b)
			sess.Flag |= SESS_KICKED_OUT
			return
		}
	} else { // agent保留协议段 [0, MAX_PROTO_NUM]
		// handle有效性检查
		h := client_handler.Handlers[b]
		if h == nil {
			log.Errorf("service id:%v not bind", b)
			sess.Flag |= SESS_KICKED_OUT
			return
		}
		// 执行
		ret = h(sess, reader)
	}

	// 统计处理时间
	elasped := time.Now().Sub(start)
	if b != 0 { // 排除心跳包日志
		log.Trace("[REQ]", b)
		_statter.Timing(1.0, fmt.Sprintf("%v%v", STATSD_PREFIX, b), elasped)
	}
}
