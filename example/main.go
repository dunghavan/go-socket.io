package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	socketio "github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"time"
)

var connArr = make(map[string]*socketio.Conn, 0)

func main() {
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext(s.ID())
		//connArr = append(connArr, &s)
		glog.Infof("1 connected at namespace=%s with URL=%s", s.Namespace(), s.URL().Path)
		s.Emit("register")
		return nil
	})
	server.OnConnect("/socket.io/gsm-server", func(s socketio.Conn) error {
		//s.SetContext(s.ID())
		glog.Infof("2 connected at namespace=%s with URL=%s, ID=%s", s.Namespace(), s.URL().Path, s.ID())
		s.Emit("register")
		return nil
	})
	server.OnEvent("/socket.io/gsm-server", "confirm", func(s socketio.Conn, msg string) {
		fmt.Println("-----------------------------------confirm:", msg)
		//s.Emit("reply", "have "+msg)
	})
	server.OnEvent("", "confirm", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		fmt.Printf("handle /chat: %s\n", msg)
		return "recv " + msg
	})
	server.OnEvent("/", "confirm", func(s socketio.Conn, msg string) {
		glog.Infof("handle confirm: %s\n", msg)
	})
	server.OnEvent("/socket.io/gsm-server", "confirm", func(s socketio.Conn, msg string) string {
		glog.Infof("handle confirm: %s\n", msg)
		connArr[msg] = &s
		return "cc"
	})
	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})
	server.OnError("/", func(e error) {
		fmt.Println("meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		fmt.Println("closed", msg)
	})
	go server.Serve()
	defer server.Close()
	go func() {
		i := 0
		for {
			time.Sleep(2 * time.Second)
			for key, conn := range connArr {
				fmt.Printf("This connection key=%s: %v\n", key, *conn)
				c := *conn
				fmt.Printf("Emit...%v\n", i)
				c.Emit("transfer_money", i)
			}

			//server.BroadcastToRoom("", "custom", "aaa")
			i++
		}
	}()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
