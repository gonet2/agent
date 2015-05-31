package main

import (
	"registry"
	. "types"
	"utils"
)

// 会话结束时的扫尾清理工作
func close_work(sess *Session) {
	defer utils.PrintPanicStack()
	if sess.Flag&SESS_LOGGED_IN == 0 {
		return
	}

	// 反注册,成功后不再接收消息
	registry.Unregister(sess.UserId)
	// 处理尚未处理的IPC数据
	len_MQ := len(sess.MQ)
	for i := 0; i < len_MQ; i++ {
		proxy_ipc_request(sess, <-sess.MQ)
	}
}
