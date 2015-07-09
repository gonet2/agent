package client_handler

import "misc/packet"
import . "types"

var Code = map[string]int16{
	"heart_beat_req":         0,  // 心跳包..
	"user_login_req":         1,  // 客户端发送用户登陆请求包
	"user_login_succeed_ack": 2,  // 登陆成功
	"user_login_faild_ack":   3,  // 登陆失败
	"client_error_ack":       4,  // 客户端错误
	"heart_beat_ack":         5,  // 心跳包回复
	"get_seed_req":           30, // socket通信加密使用
	"get_seed_ack":           31, // socket通信加密使用
	"game_servers_req":       32, // 拉取服务器列表
	"game_servers_ack":       33, // 拉取服务器列表回复
	"choose_server_req":      34, // 选择服务器
}

var RCode = map[int16]string{
	0:  "heart_beat_req",         // 心跳包..
	1:  "user_login_req",         // 客户端发送用户登陆请求包
	2:  "user_login_succeed_ack", // 登陆成功
	3:  "user_login_faild_ack",   // 登陆失败
	4:  "client_error_ack",       // 客户端错误
	5:  "heart_beat_ack",         // 心跳包回复
	30: "get_seed_req",           // socket通信加密使用
	31: "get_seed_ack",           // socket通信加密使用
	32: "game_servers_req",       // 拉取服务器列表
	33: "game_servers_ack",       // 拉取服务器列表回复
	34: "choose_server_req",      // 选择服务器
}

var Handlers map[int16]func(*Session, *packet.Packet) []byte

func init() {
	Handlers = map[int16]func(*Session, *packet.Packet) []byte{
		0:  P_heart_beat_req,
		1:  P_user_login_req,
		30: P_get_seed_req,
		32: P_game_servers_req,
		34: P_choose_server_req,
	}
}
