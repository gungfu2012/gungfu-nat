package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
)

var remoteserver string = "ws://127.0.0.1:8080/"

const ctlpath string = "control"
const bufmax uint = 1 << 20

func readfromconn(conn net.Conn, wsconn *websocket.Conn) {
	var buf [bufmax]byte
	for {
		if conn == nil || wsconn == nil {
			break
		}
		n, err := conn.Read(buf[0:bufmax])
		if err != nil && n == 0 {
			conn.Close()
			break
		}
		if n == 0 {
			continue
		}
		err = wsconn.WriteMessage(websocket.BinaryMessage, buf[0:n])
		if err != nil {
			wsconn.Close()
			break
		}
	}
}

func writetoconn(conn net.Conn, wsconn *websocket.Conn) {
	for {
		if conn == nil || wsconn == nil {
			break
		}
		_, buf, err := wsconn.ReadMessage()
		if err != nil {
			wsconn.Close()
			break
		}
		_, err = conn.Write(buf)
		if err != nil {
			conn.Close()
			break
		}
	}
}
func main() {
	var port, path string
	flag.Parse()

	var index int = 0
	var header = http.Header{}
	header.Add("conn-index", strconv.Itoa(index))

	ctlconn, _, _ := websocket.DefaultDialer.Dial(remoteserver+ctlpath, nil)
	fmt.Println("create ctl connection")

	for {
		mt, buf, _ := ctlconn.ReadMessage()
		if mt != websocket.BinaryMessage {
			continue
		}
		switch buf[0] {
		case 0: //ssh
			path = "ssh_localserver"
			port = "22"
		case 1: //emby
			path = "emby_localserver"
			port = "8096"
		}
		header.Set("conn-index", string(buf[2:2+buf[1]]))
		wsconn, _, _ := websocket.DefaultDialer.Dial(remoteserver+path, header)
		conn, _ := net.Dial("tcp", "127.0.0.1:"+port)
		go readfromconn(conn, wsconn)
		go writetoconn(conn, wsconn)
	}
}
