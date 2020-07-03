package common

import "strconv"
import "os"
import (
	"math/rand"
	"time"
)

// ** adapted from mapreduce
func GameServerSock() string {
	s := "/var/tmp/824-gs-"
	s += strconv.Itoa(os.Getuid())
	return s
}

type Card struct {
	Value string
	Suit  string
}

type Pairs struct {
	One Card
	Two Card
}

type Player struct {
	ID int
	Hand []Card
	Pairs []Pairs
}

type JoinGameArgs struct {

}

type JoinGameReply struct {
	Success bool
	ID int
	// Hand  []Card
	// Pairs []Pairs
}

type GameStatusRequest struct {
	
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
