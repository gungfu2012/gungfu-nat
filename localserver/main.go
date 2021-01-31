package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"strconv"
	"time"
)

const ctlpath string = "control"
const bufmax uint = 1 << 20
var ctlconn *websocket.Conn

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

func sendping() {
	var errcount int = 0
	for {
		if ctlconn == nil {
			break
		}
		err := ctlconn.WriteMessage(websocket.PingMessage, nil)
		if err != nil {
			errcount ++
			fmt.Println(err)
			if errcount > 5 {
				break
			}
			continue
		}
		time.Sleep(10 * time.Second)
		errcount = 0
	}
}

func main() {
	var port, path, remoteserver string
	flag.StringVar(&remoteserver,"remoteserver","ws://127.0.0.1:8080/","default remote server")
	flag.Parse()
	for {
		var index int = 0
		var header = http.Header{}
		header.Add("conn-index", strconv.Itoa(index))

		ctlconn, _, err := websocket.DefaultDialer.Dial(remoteserver+ctlpath, nil)
		if err != nil {
			fmt.Println(err)
			time.Sleep(60 * time.Second)
			continue
		}
		go sendping()
		fmt.Println("create ctl connection")

		var errcount int =0
		for {
			if ctlconn == nil {
				break
			}
			mt, buf, err := ctlconn.ReadMessage()
			if err != nil {
				errcount ++
				fmt.Println(err)
				if errcount > 5 {
					break
				}
				continue
			}
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
			wsconn, _, err := websocket.DefaultDialer.Dial(remoteserver+path, header)
			if err != nil {
				continue
			}
			conn, err := net.Dial("tcp", "127.0.0.1:"+port)
			if err != nil {
				continue
			}
			go readfromconn(conn, wsconn)
			go writetoconn(conn, wsconn)
			errcount = 0
		}
	}
}
