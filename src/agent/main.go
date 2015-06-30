package main

import (
	"encoding/binary"
	log "github.com/GameGophers/nsq-logger"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

import (
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

	// create a write buffer
	out := new_buffer(conn, sess_die)
	go out.start()

	// start one agent for handling packet
	wg.Add(1)
	go agent(&sess, in, out, sess_die)

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
		payload := make([]byte, size)
		// read msg
		n, err = io.ReadFull(conn, payload)
		if err != nil {
			log.Warningf("read payload failed, ip:%v reason:%v size:%v", sess.IP, err, n)
			return
		}

		select {
		case in <- payload: // payload queued
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
