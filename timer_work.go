package main

import (
	log "github.com/Sirupsen/logrus"
)

import (
	. "agent/types"
)

var (
	rpmLimit int
)

func initTimer(rpm_limit int) {
	rpmLimit = rpm_limit
}

// 玩家1分钟定时器
func timer_work(sess *Session, out *Buffer) {
	defer func() {
		sess.PacketCount1Min = 0
	}()

	// 发包频率控制，太高的RPS直接踢掉
	if sess.PacketCount1Min > rpmLimit {
		sess.Flag |= SESS_KICKED_OUT
		log.WithFields(log.Fields{
			"userid":  sess.UserId,
			"count1m": sess.PacketCount1Min,
			"total":   sess.PacketCount,
		}).Error("RPM")
		return
	}
}
