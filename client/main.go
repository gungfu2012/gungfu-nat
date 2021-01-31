package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
)


const bufmax uint = 1 << 20

func readfromconn(conn net.Conn, wsconn *websocket.Conn) {
	var buf [bufmax]byte
	for {
		if conn == nil || wsconn == nil {
			break
		}
		n, err := conn.Read(buf[0:bufmax])
		if err != nil && n == 0 {
			fmt.Println("read from local err:",err)
			break
		}
		if n == 0 {
			continue
		}
		fmt.Println("read data from local")
		err = wsconn.WriteMessage(websocket.BinaryMessage, buf[0:n])
		if err != nil {
			fmt.Println("write data to remote err:",err)
			break
		}
		fmt.Println("write data to remote")
	}
}

func writetoconn(conn net.Conn, wsconn *websocket.Conn) {
	for {
		if conn == nil || wsconn == nil {
			break
		}
		_, buf, err := wsconn.ReadMessage()
		if err != nil {
			fmt.Println("read data from remote err:",err)
			break
		}
		fmt.Println("read data from remote")
		_, err = conn.Write(buf)
		if err != nil {
			fmt.Println("write data to local err:",err)
			break
		}
		fmt.Println("write data to local")
	}
}
func main() {
	var port, path, remoteserver string
	flag.StringVar(&port, "port", "2222", "default port for ssh")
	flag.StringVar(&path, "path", "ssh_client", "default path for ssh")
	flag.StringVar(&remoteserver,"remoteserver","ws://127.0.0.1:8080/","default remote server")
	flag.Parse()

	var index int = 0
	var header = http.Header{}
	header.Add("conn-index", strconv.Itoa(index))

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("start listen the ", port, " port")
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		wsconn, _, err := websocket.DefaultDialer.Dial(remoteserver+path, header)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("get a connction,the index :",index)
		go readfromconn(conn, wsconn)
		go writetoconn(conn, wsconn)

		index = (index + 1) % 65536
		header.Set("conn-index", strconv.Itoa(index))
	}
}
