package p2p

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Aergiaaa/poker/deck"
)

type GameStatus int32

const (
	GameStatusWaitingForCards GameStatus = iota
	GamseStatusReceivingCards
	GameStatusDealing
	GameStatusPreflop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

func (g GameStatus) string() string {
	switch g {
	case GameStatusWaitingForCards:
		return "Waiting for Cards"
	case GamseStatusReceivingCards:
		return "Receiving Cards"
	case GameStatusDealing:
		return "Dealing"
	case GameStatusPreflop:
		return "Pre flop"
	case GameStatusFlop:
		return "Flop"
	case GameStatusTurn:
		return "Turn"
	case GameStatusRiver:
		return "River"
	default:
		return "UNKNOWN STATUS"
	}
}

type Player struct {
	Status GameStatus
}

type GameState struct {
	broadcastch chan any
	listenAddr  string
	isDealer    bool

	gameStatus GameStatus

	playersWaitingForCards int32
	playersLock            sync.RWMutex
	players                map[string]*Player
}

func NewGameState(addr string, broadcastch chan any) *GameState {
	g := &GameState{
		broadcastch: broadcastch,
		listenAddr:  addr,
		isDealer:    false,
		gameStatus:  GameStatusWaitingForCards,
		players:     make(map[string]*Player),
	}

	go g.loop()

	return g
}

// TODO: check other RW occurence of the GameStatus
func (g *GameState) SetStatus(s GameStatus) {
	atomic.StoreInt32((*int32)(&g.gameStatus), int32(s))
}

func (g *GameState) AddPlayerWaitingForCards() {
	atomic.AddInt32(&g.playersWaitingForCards, 1)
}

func (g *GameState) CheckNeedsDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaitingForCards)

	condition := g.isDealer &&
		g.gameStatus == GameStatusWaitingForCards &&
		playersWaiting == int32(len(g.players))

	if condition {
		fmt.Printf("[%s]dealing cards!!!\n", g.listenAddr)

		g.DealCards()
	}

}

func (g *GameState) DealCards() {
	g.broadcastch <- MessageCards{Deck: deck.New()}
}

func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
	players, ok := g.players[addr]
	if !ok {
		panic("player could not be found, altough it should exist")
	}

	players.Status = status

	g.CheckNeedsDealCards()
}

func (g *GameState) LenPlayersConnectedWithLock() int {
	g.playersLock.RLock()
	defer g.playersLock.RUnlock()

	return len(g.players)
}

func (g *GameState) AddPlayer(addr string, status GameStatus) {
	g.playersLock.Lock()
	defer g.playersLock.Unlock()

	if status == GameStatusWaitingForCards {
		g.AddPlayerWaitingForCards()
	}

	g.players[addr] = new(Player)

	// set the player status also when we add the player
	g.SetPlayerStatus(addr, status)

	fmt.Printf("New Player Joined [%s][%s]\n", addr, status.string())
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-ticker.C:
			fmt.Printf("player connected [%d][%s]\n", g.LenPlayersConnectedWithLock(), g.gameStatus.string())
		}
	}
}
