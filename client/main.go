package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
)

var remoteserver string = "ws://192.168.1.168:8080/"

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
	flag.StringVar(&port, "port", "2222", "default port for ssh")
	flag.StringVar(&path, "path", "ssh_client", "default path for ssh")
	flag.Parse()

	var index int = 0
	var header = http.Header{}
	header.Add("conn-index", strconv.Itoa(index))

	ln, err := net.Listen("tcp", ":"+port)
	fmt.Println(err)
	fmt.Println("start listen the ", port, " port")
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		wsconn, _, err := websocket.DefaultDialer.Dial(remoteserver+path, header)
		if err != nil {
			continue
		}
		go readfromconn(conn, wsconn)
		go writetoconn(conn, wsconn)

		index = (index + 1) % 65536
		header.Set("conn-index", strconv.Itoa(index))
	}
}
