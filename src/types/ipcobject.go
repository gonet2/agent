package types

// IPCObject 定义
type IPCObject struct {
	Id      uint64      `msgpack:"id"` // 消息ID(用于幂等操作)
	SrcId   int32       `msgpack:"a"`  // 发送方用户ID
	DestId  int32       `msgpack:"b"`  // 接收放用户ID
	AuxIds  []int32     `msgpack:"c"`  // 目标用户ID集合(用于组播)
	Service int16       `msgpack:"d"`  // 服务号
	Object  interface{} `msgpack:"e"`  // 投递的内容
	Time    int64       `msgpack:"f"`  // 发送时间
	Extra   []byte      `msgpack:"g"`  // 额外原始数据
}
