package main

import (
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	log "github.com/GameGophers/libs/nsq-logger"
	"github.com/GameGophers/libs/services"
	pb "github.com/GameGophers/libs/services/proto"
	_ "github.com/GameGophers/libs/statsd-pprof"
)

import (
	_ "ipc"
	_ "numbers"
	. "types"
	"utils"
)

const (
	_port = ":8888"
)

const (
	SERVICE = "[AGENT]"
)

func main() {
	defer utils.PrintPanicStack()
	go func() {
		log.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	log.SetPrefix(SERVICE)

	// server startup procedure
	startup()

	tcpAddr, err := net.ResolveTCPAddr("tcp4", _port)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	log.Info("listening on:", listener.Addr())

	// loop accepting
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Warning("accept failed:", err)
			continue
		}
		go handleClient(conn)

		// check server close signal
		select {
		case <-die:
			listener.Close()
			goto FINAL
		default:
		}
	}
FINAL:
	// server closed, wait forever
	for {
		<-time.After(time.Second)
	}
}

// start a goroutine when a new connection is accepted
func handleClient(conn *net.TCPConn) {
	defer utils.PrintPanicStack()
	// set per-connection socket buffer
	conn.SetReadBuffer(SO_RCVBUF)

	// set initial socket buffer
	conn.SetWriteBuffer(SO_SNDBUF)

	// initial network control struct
	header := make([]byte, 2)
	in := make(chan []byte)
	defer func() {
		close(in) // session will close
	}()

	// pre-allocated packet buffer for each connection
	prealloc_buf := make([]byte, PREALLOC_BUFSIZE)
	index := 0

	// create a new session object for the connection
	var sess Session
	host, port, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Error("cannot get remote address:", err)
		return
	}
	sess.IP = net.ParseIP(host)
	log.Infof("new connection from:%v port:%v", host, port)

	// session die signal
	sess_die := make(chan bool)

	//connect to game service and recv data
	cli, err := services.GetService(services.SERVICE_GAME)
	if err != nil {
		log.Critical(err)
		return
	}
	service, _ := cli.(pb.GameServiceClient)
	stream, err := service.Packet(context.Background())
	if err != nil {
		log.Critical(err)
		return
	}
	defer stream.CloseSend()

	service_chan := make(chan *pb.Game_Packet, PREALLOC_BUFSIZE)
	go game_service(stream, service_chan)

	// create a write buffer
	out := new_buffer(conn, sess_die)
	go out.start()

	// start one agent for handling packet
	wg.Add(1)
	go agent(&sess, in, out, stream, service_chan, sess_die)

	// network loop
	for {
		// solve dead link problem
		conn.SetReadDeadline(time.Now().Add(TCP_READ_DEADLINE * time.Second))
		n, err := io.ReadFull(conn, header)
		if err != nil {
			log.Warningf("read header failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}
		size := binary.BigEndian.Uint16(header)

		// alloc a byte slice for reading
		if index+int(size) > PREALLOC_BUFSIZE {
			index = 0
			prealloc_buf = make([]byte, PREALLOC_BUFSIZE)
		}
		data := prealloc_buf[index : index+int(size)]
		index += int(size)

		// read msg
		n, err = io.ReadFull(conn, data)
		if err != nil {
			log.Warningf("read msg failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}

		select {
		case in <- data: // data queued
		case <-sess_die:
			log.Warningf("connection closed by logic, flag:%v ip:%v", sess.Flag, sess.IP)
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Critical(err)
		os.Exit(-1)
	}
}
