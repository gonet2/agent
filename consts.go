package main

const (
	// 网络拥塞控制和削峰
	readDeadline  = 15       // 秒(没有网络包进入的最大间隔)
	receiveBuffer = 32767    // 每个连接的接收缓冲区
	sendBuffer    = 65535    // 每个连接的发送缓冲区
	udpBuffer     = 16777216 // UDP监听器的socket buffer
	tosEF         = 46       // Expedited Forwarding (EF)
)

const (
	rpmLimit = 200 // Request Per Minute
)
