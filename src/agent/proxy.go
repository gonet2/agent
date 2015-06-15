package main

import (
	"os"
	"time"

	log "github.com/GameGophers/nsq-logger"
	"github.com/peterbourgon/g2s"
)

import (
	"client_handler"
	"ipc_handler"
	"misc/packet"
	. "types"
	"utils"
)

const (
	STATSD_PREFIX       = "API."
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
func proxy_user_request(sess *Session, p []byte) []byte {
	defer utils.PrintPanicStack(sess, p)
	// 解密
	if sess.Flag&SESS_ENCRYPT != 0 {
		sess.Decoder.Codec(p)
	}

	// 封装为reader
	reader := packet.Reader(p)

	// 读客户端数据包序列号(1,2,3...)
	// 可避免重放攻击-REPLAY-ATTACK
	seq_id, err := reader.ReadU32()
	if err != nil {
		log.Error("read client timestamp failed:", err)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// 读协议号
	b, err := reader.ReadS16()
	if err != nil {
		log.Error("read protocol number failed.")
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// 数据包序列号验证
	if seq_id != sess.PacketCount {
		log.Errorf("illegal packet sequence id:%v should be:%v proto:%v size:%v", seq_id, sess.PacketCount, b, len(p)-6)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// handle有效性检查
	h := client_handler.Handlers[b]
	if h == nil {
		log.Errorf("service id:%v not bind", b)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// 前置HOOK
	if !_before_hook(sess, b) {
		log.Error("before hook failed, code:", b)
		sess.Flag |= SESS_KICKED_OUT
		return nil
	}

	// 包处理
	start := time.Now()
	ret := h(sess, reader)
	end := time.Now()

	// 打印协议，登陆前的协议不会打印
	if sess.Flag&SESS_LOGGED_IN != 0 {
		if b != 0 { // 排除心跳包日志
			log.Trace("[REQ]", client_handler.RCode[b])
			_statter.Timing(1.0, STATSD_PREFIX+client_handler.RCode[b], end.Sub(start))
		}
	}

	// 后置HOOK
	_after_hook(sess, b)

	return ret
}

// IPC proxy
func proxy_ipc_request(sess *Session, p IPCObject) []byte {
	defer utils.PrintPanicStack()
	h := ipc_handler.Handlers[p.Service]

	// 获取Handler
	if h == nil {
		log.Errorf("ipc service: %v not bind, internal service error.", p.Service)
		return nil
	}

	// IPCObject处理
	// start := time.Now()
	ret := h(sess, p)
	// end := time.Now()
	// log.Printf("\033[0;36m[IPC] %v\t%v\033[0m\n", p.Service, end.Sub(start))
	return ret
}

// 前置HOOK
func _before_hook(sess *Session, rcode int16) bool {
	if sess.Flag&SESS_LOGGED_IN == 0 {
		return true
	}

	return true
}

// 后置HOOK
func _after_hook(sess *Session, rcode int16) {
	if sess.Flag&SESS_LOGGED_IN == 0 {
		return
	}

	//check need flush to db or not
	flush(sess, rcode)
}
