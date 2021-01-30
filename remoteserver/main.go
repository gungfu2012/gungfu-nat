package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

const bufmax uint = 1 << 20

const arraycount int = 65536

var ctlconn *websocket.Conn

var ssh_client_conn [arraycount]*websocket.Conn
var ssh_localserver_conn [arraycount]*websocket.Conn
var emby_client_conn [arraycount]*websocket.Conn
var emby_localserver_conn [arraycount]*websocket.Conn

var upgrader = websocket.Upgrader{}

func control(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	ctlconn = c
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ssh_client(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	index := r.Header.Get("conn-index")
	indexlen := len(index)
	index_int, err := strconv.Atoi(index)
	if err != nil {
		return
	}
	var buf [258]byte
	buf[0] = 0x00
	buf[1] = uint8(indexlen)
	for i := 2; i < indexlen+2; i++ {
		buf[i] = index[i-2]
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	ssh_client_conn[index_int] = c
	if ctlconn == nil {
		return
	}
	err = ctlconn.WriteMessage(websocket.BinaryMessage, buf[0:indexlen+2])
	if err != nil {
		fmt.Println(err)
	}
}

func ssh_localserver(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	index := r.Header.Get("conn-index")
	index_int, err := strconv.Atoi(index)
	if err != nil {
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	ssh_localserver_conn[index_int] = c
	go tunnel(ssh_localserver_conn[index_int], ssh_client_conn[index_int])
	go tunnel(ssh_client_conn[index_int], ssh_localserver_conn[index_int])
}

func emby_client(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	index := r.Header.Get("conn-index")
	indexlen := len(index)
	index_int, err := strconv.Atoi(index)
	if err != nil {
		return
	}
	var buf [258]byte
	buf[0] = 0x01
	buf[1] = uint8(indexlen)
	for i := 2; i < indexlen+2; i++ {
		buf[i] = index[i-2]
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	emby_client_conn[index_int] = c
	if ctlconn == nil {
		return
	}
	err = ctlconn.WriteMessage(websocket.BinaryMessage, buf[0:indexlen+2])
	if err != nil {
		fmt.Println(err)
	}
}

func emby_localserver(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	index := r.Header.Get("conn-index")
	index_int, err := strconv.Atoi(index)
	if err != nil {
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	emby_localserver_conn[index_int] = c
	go tunnel(emby_localserver_conn[index_int], emby_client_conn[index_int])
	go tunnel(emby_client_conn[index_int], emby_localserver_conn[index_int])

}

func tunnel(r, w *websocket.Conn) {
	for {
		if r == nil || w == nil {
			return
		}
		mt, buf, err := r.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		err = w.WriteMessage(mt, buf)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/control", control)
	http.HandleFunc("/ssh_client", ssh_client)
	http.HandleFunc("/ssh_localserver", ssh_localserver)
	http.HandleFunc("/emby_client", emby_client)
	http.HandleFunc("/emby_localserver", emby_localserver)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	fmt.Println("this server port is : ",port)
	addr := ":4000"
	fmt.Println("this server addr is : ",addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
