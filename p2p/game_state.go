package p2p

type GameStatus uint32

const (
	GameStatusWaiting GameStatus = iota
	GameStatusDealing
	GameStatusPreflop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

func (g GameStatus) string() string {
	switch g {
	case GameStatusWaiting:
		return "Waiting"
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
		return "Unknown Status"
	}
}

type GameState struct {
	gameStatus GameStatus
	isDealer   bool
}

func NewGameState() *GameState {
	return &GameState{}
}

func (g *GameState) loop() {
	for {
		select {}
	}
}
