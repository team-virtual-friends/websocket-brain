package main

import (
	"log"

	"github.com/lesismal/nbio"
)

func main() {
	engine := nbio.NewEngine(nbio.Config{
		Network:            "tcp", //"udp", "unix"
		Addrs:              []string{":8510"},
		MaxWriteBufferSize: 6 * 1024 * 1024,
	})

	// hanlde new connection
	engine.OnOpen(func(c *nbio.Conn) {
		log.Println("OnOpen:", c.RemoteAddr().String())
	})
	// hanlde connection closed
	engine.OnClose(func(c *nbio.Conn, err error) {
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})
	// handle data
	engine.OnData(func(c *nbio.Conn, data []byte) {
		c.Write(append([]byte{}, data...))
	})

	err := engine.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
		return
	}
	defer engine.Stop()

	<-make(chan int)
}
