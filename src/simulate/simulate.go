package main

import (
	"log"
	"misc/packet"
	"net"
	"os"
	"time"
)

var seqid = uint32(1)

func checkErr(err error) {
	if err != nil {
		log.Println(err)
		panic("error occured in protocol module")
	}
}
func main() {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
	defer conn.Close()

	//send login
	info := &user_login_info{
		F_open_udid: "Robot agent",
		F_login_way: 1,
	}
	send_proto(conn, Code["user_login_req"], info)
}

func send_proto(conn net.Conn, p int16, info interface{}) {
	seqid++
	payload := packet.Pack(p, info, nil)
	writer := packet.Writer()
	writer.WriteU16(uint16(len(payload)) + 4)
	writer.WriteU32(seqid)
	writer.WriteRawBytes(payload)
	conn.Write(writer.Data())
	time.Sleep(time.Second)
}
