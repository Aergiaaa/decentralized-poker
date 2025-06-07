package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"slices"
	"sync"
	"time"
)

type GameType byte

const (
	Texas GameType = iota
	Poker
)

func (gc GameType) String() string {
	switch gc {
	case Texas:
		return "Texas"
	case Poker:
		return "Poker"
	default:
		return "UNKNOWN"
	}
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	GameType
}

type Server struct {
	ServerConfig
	transport   *TCPTransport
	gameState   *GameState
	peerLock    sync.RWMutex
	peers       map[net.Addr]*Peer
	addPeer     chan *Peer
	delPeer     chan *Peer
	msgCh       chan *Message
	broadcastCh chan any
}

func NewServer(cfg ServerConfig, connection int) *Server {
	s := &Server{
		ServerConfig: cfg,
		peers:        make(map[net.Addr]*Peer),
		addPeer:      make(chan *Peer, connection),
		delPeer:      make(chan *Peer, connection),
		msgCh:        make(chan *Message),
		broadcastCh:  make(chan any),
	}
	s.gameState = NewGameState(s.ListenAddr, s.broadcastCh)

	if s.ListenAddr == ":3000" {
		s.gameState.isDealer = true //for testing
	}

	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr

	tr.AddPeer, tr.DelPeer = s.addPeer, s.delPeer

	return s
}

func (s *Server) Start() {
	go s.loop()

	fmt.Printf("Started new game server port %s game :%s \n", s.ListenAddr, s.GameType.String())

	// TODO: NEED TO BE MOVED LATER (ugly)
	go func() {
		msg := <-s.broadcastCh
		fmt.Printf("broadcasting!!!\n")
		if err := s.Broadcast(msg); err != nil {
			fmt.Printf("error broadcasting: %s\n", err)
		}
	}()

	if err := s.transport.ListenAndAccept(); err != nil {
		panic(err)
	}
}

func (s *Server) sendPeerList(p *Peer) error {
	peerList := MessagePeerList{
		Peers: []string{},
	}

	peers := s.Peers()
	for i := range peers {
		if peers[i] != p.listenAddr {
			peerList.Peers = append(peerList.Peers, peers[i])
		}
	}

	if len(peerList.Peers) == 0 {
		return nil
	}

	msg := NewMessage(s.ListenAddr, peerList)
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) sendHandshake(p *Peer) error {
	hs := &Handshake{
		GameType:   s.GameType,
		Version:    s.Version,
		GameStatus: s.gameState.gameStatus,
		ListenAddr: s.ListenAddr,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}
	return p.Send(buf.Bytes())
}

func (s *Server) Connect(addr string) error {
	if s.isInPeerList(addr) {
		return nil
	}

	conn, err := net.DialTimeout("tcp", addr, time.Second*1)
	if err != nil {
		return err
	}

	peer := &Peer{
		conn:     conn,
		outbound: true,
	}

	s.addPeer <- peer

	return s.sendHandshake(peer)
}

func (s *Server) loop() {
	for {
		select {
		// case msg := <-s.broadcastCh:
		// 	fmt.Printf("broadcasting!!!\n")
		// 	if err := s.Broadcast(msg); err != nil {
		// 		fmt.Printf("error broadcasting: %s\n", err)
		// 	}
		case peer := <-s.delPeer:
			delete(s.peers, peer.conn.RemoteAddr())
			fmt.Printf("[%s] peer %s disconnected\n", s.ListenAddr, peer.conn.RemoteAddr())

		case peer := <-s.addPeer:
			if err := s.handleNewPeer(peer); err != nil {
				fmt.Printf("[%s] error handling new peer: %s", s.ListenAddr, err)
			}

		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				fmt.Printf("[%s] error handling msg: %s", s.ListenAddr, err)
			}
		}
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	hs, err := s.handshake(peer)
	if err != nil {
		peer.conn.Close()
		delete(s.peers, peer.conn.RemoteAddr())

		return fmt.Errorf("handshake with incoming peer failed, %s\n", err)
	}

	go peer.ReadLoop(s.msgCh)

	if !peer.outbound {
		err := s.sendHandshake(peer)
		if err != nil {
			peer.conn.Close()
			delete(s.peers, peer.conn.RemoteAddr())

			return fmt.Errorf("handshake with incoming peer failed, %s\n", err)
		}

		if err := s.sendPeerList(peer); err != nil {
			return fmt.Errorf("Peer List error: %s\n", err)
		}
	}

	fmt.Printf("[%s] Handshake successful with [%s]: %s %s, status:%s\n",
		s.ListenAddr, peer.listenAddr,
		hs.Version, hs.GameType, hs.GameStatus.string(),
	)

	s.handleAddPeer(peer)

	s.gameState.AddPlayer(peer.listenAddr, hs.GameStatus)

	return nil
}

func (s *Server) handleAddPeer(p *Peer) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.conn.RemoteAddr()] = p
}

func (s *Server) Peers() []string {
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()

	peers := make([]string, len(s.peers))
	iter := 0
	for _, p := range s.peers {
		peers[iter] = p.listenAddr
		iter++
	}

	return peers
}

func (s *Server) Broadcast(payload any) error {
	var err error

	msg := NewMessage(s.ListenAddr, payload)

	buf := new(bytes.Buffer)
	if err = gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, p := range s.peers {
		go func(p *Peer) {
			if err = p.Send(buf.Bytes()); err != nil {
				fmt.Println("Broadcast error: ", err)
			}
			fmt.Printf("[%s]sending cards to [%s]\n", s.ListenAddr, p.listenAddr)
		}(p)

	}

	return nil
}

func (s *Server) handshake(p *Peer) (*Handshake, error) {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return nil, err
	}

	if s.GameType != hs.GameType {
		return nil, fmt.Errorf("different gametype: %s", hs.GameType)
	}
	if s.Version != hs.Version {
		return nil, fmt.Errorf("different version: %s", hs.Version)
	}

	p.listenAddr = hs.ListenAddr

	return hs, nil
}

func (s *Server) handleMessage(msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessagePeerList:
		return s.handlePeerList(v)
	case MessageCards:
		fmt.Printf("[%s]received msg from [%s] -> %+v\n", s.ListenAddr, msg.From, v)

		// TODO: figuring out the status when we are receiving cards
		s.gameState.SetStatus(GamseStatusReceivingCards)
	}

	return nil
}

func (s *Server) handlePeerList(pl MessagePeerList) error {
	p := pl.Peers
	for i := range p {
		if err := s.Connect(p[i]); err != nil {
			fmt.Printf("[%s] Error failed to dial peer [%s]: %s\n", s.ListenAddr, p[i], err)
			continue
		}
	}
	return nil
}

func (s *Server) isInPeerList(addr string) bool {
	peers := s.Peers()

	return slices.Contains(peers, addr)

}

func init() {

	register := []any{
		MessagePeerList{},
		MessageCards{},
	}

	for _, v := range register {
		gob.Register(v)
	}
}
