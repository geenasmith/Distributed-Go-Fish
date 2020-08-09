package helpers

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

type GameStateArgs struct {
	Key     string
	Payload string
	Ok      bool
}
type GameStateReply struct {
	Payload string
	Ok      bool
}

type Card struct {
	Value string
	Suit  string
	Used  bool // has been played and discarded
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
func RaftServerSock() string {
	s := "/var/tmp/824-rb-"
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

// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func CallGS(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := GameServerSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func CallRB(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := RaftServerSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
