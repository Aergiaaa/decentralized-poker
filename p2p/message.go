package p2p

import "github.com/Aergiaaa/poker/deck"

type Message struct {
	Payload any
	From    string
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		Payload: payload,
		From:    from,
	}
}

type Handshake struct {
	GameType
	GameStatus
	Version    string
	ListenAddr string
}

type MessagePeerList struct {
	Peers []string
}

type MessageCards struct {
	Deck deck.Deck
}
