package main

const (
	// 网络拥塞控制和削峰
	TCP_READ_DEADLINE = 120   // 秒(没有网络包进入的最大间隔)
	SO_RCVBUF         = 32767 // 每个连接的接收缓冲区
	SO_SNDBUF         = 65535 // 每个连接的发送缓冲区
)

const (
	PADDING_LIMIT         = 8   // 小于此的返回包，加入填充
	PADDING_SIZE          = 8   // 填充最大字节数
	PADDING_UPDATE_PERIOD = 300 // 填充字符更新周期
)

const (
	DEFAULT_MQ_SIZE  = 512   // 默认玩家IPC消息队列大小
	CUSTOM_TIMER     = 60    // 玩家定时器间隔
	PREALLOC_BUFSIZE = 65536 // 预分配的接收缓冲
)

const (
	SYS_TIMER   = 600 // 系统进程定时任务间隔(s)
	GC_INTERVAL = 300 // 主动垃圾回收间隔
	RPM_LIMIT   = 300 // Request Per Minute
)
