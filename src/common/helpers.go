package common

import (
	"math/rand"
	"time"
)

type Card struct {
	Value string
	Suit  string
}

type Pairs struct {
	one Card
	two Card
}

type Player struct {
	Hand  []Card
	Pairs []Pairs
}

type GameStatusReply struct {
	Complete      bool
	Turn          int
	CurrentPlayer int
	Winner        int
	Scores        []int
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
	Pair  []Pairs
}

type PlayPairReply struct {
	Turn     int
	Accepted bool
}

func Shuffle(slice []Card) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIdx := r.Intn(n)
		slice[n-1], slice[randIdx] = slice[randIdx], slice[n-1]
	}
}
