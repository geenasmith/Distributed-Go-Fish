package common

import "strconv"
import "os"

// ** adapted from mapreduce
func GameServerSock() string {
	s := "/var/tmp/824-gs-"
	s += strconv.Itoa(os.Getuid())
	return s
}

type Card struct {
	Value string
	Suit string
}

type Pairs struct {
	one Card
	two Card
}

type Player struct {
	Hand []Card
	Pairs []Pairs
}

type JoinGameArgs struct {

}

type JoinGameReply struct {
	Success bool
}
