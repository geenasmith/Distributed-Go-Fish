package main

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

// Struct to hold the state of the game
type GameServer struct {
	Mu                sync.Mutex
	Ready             bool
	GameOver          bool
	Winner            int // index of the winning player
	Players           []Player // holds ID, hand, pairs, and opponents
	Deck              []Card   // hold the cards that are still in the deck
	CurrentTurnPlayer int // what's the difference between this and currentTurn???
	CurrentTurn       int
	PlayerCount       int // number of players in the game
	GameInitialized   bool
}

type Player struct {
	ID        int
	Hand      []Card
	Pairs     []Pairs
	Opponents []Player
}

type Card struct {
	Value string
	Suit  string
	used  bool // has been played and discarded
}

type Pairs struct {
	One Card
	Two Card
}

type JoinGameArgs struct {
}

type JoinGameReply struct {
	Success bool
	ID      int
}

type GameStatusRequest struct {
}

type GameStatusReply struct {
	Complete      bool
	Turn          int
	CurrentPlayer int
	Winner        int
	Scores        []int
	Players       []Player
}

type CardRequest struct {
	Turn   int
	Target int //Index of target player
	Value  string
}

type CardRequestReply struct {
	Turn   int
	Cards  []Card
	GoFish bool
}

type PlayPairRequest struct {
	Turn  int
	Owner int
	Hand  []Card
	Pair  []Pairs
}

type PlayPairReply struct {
	Turn     int
	Accepted bool
}

func GameServerSock() string {
	s := "/var/tmp/824-gs-"
	s += strconv.Itoa(os.Getuid())
	return s
}

func Shuffle(slice []Card) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIdx := r.Intn(n)
		slice[n-1], slice[randIdx] = slice[randIdx], slice[n-1]
	}
}