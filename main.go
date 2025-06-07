package main

import (
	"time"

	"github.com/Aergiaaa/poker/p2p"
)

func makeServerAndStart(addr string, connection int) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version:    "POKER v0.1-alpha",
		ListenAddr: addr,
		GameType:   p2p.Texas,
	}
	s := p2p.NewServer(cfg, connection)

	go s.Start()

	return s
}

func wait(t time.Duration) {
	time.Sleep(time.Millisecond * t)
}

func main() {

	c := 4
	s0 := makeServerAndStart(":3000", c)
	s1 := makeServerAndStart(":3001", c)
	s2 := makeServerAndStart(":3002", c)
	s3 := makeServerAndStart(":3003", c)

	wait(200)
	var err error
	err = s2.Connect(s1.ListenAddr)
	if err != nil {
		panic(err)
	}
	wait(100)
	err = s3.Connect(s2.ListenAddr)
	if err != nil {
		panic(err)
	}
	wait(100)
	err = s0.Connect(s1.ListenAddr)
	if err != nil {
		panic(err)
	}

	select {}
}
