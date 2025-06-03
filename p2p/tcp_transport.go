package p2p

import (
	"encoding/gob"
	"fmt"
	"net"
)

type Peer struct {
	conn       net.Conn
	outbound   bool
	listenAddr string
}

func (p *Peer) Send(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *Peer) ReadLoop(msgch chan *Message) {
	for {
		msg := &Message{}
		if err := gob.NewDecoder(p.conn).Decode(msg); err != nil {
			fmt.Printf("[%s] Decoding error: %s\n", p.conn.RemoteAddr(), err)
			break
		}

		msgch <- msg

	}
	p.conn.Close()
}

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	AddPeer    chan *Peer
	DelPeer    chan *Peer
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		peer := &Peer{
			conn:     conn,
			outbound: false,
		}

		t.AddPeer <- peer
	}

}
