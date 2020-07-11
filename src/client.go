package main

import (
	"math/rand"
	"net/rpc"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Player struct {
	ID        int
	Hand      []Card
	Pairs     []Pairs
	Opponents []Player
}

// ** adapted from mapreduce
func GameServerSock() string {
	s := "/var/tmp/824-gs-"
	s += strconv.Itoa(os.Getuid())
	return s
}

type Card struct {
	Value string
	Suit  string
	Used  bool
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

func (p *Player) PlayGoFish() {
	var gameOver = false
	for !gameOver {
		var reply = callGetGameStatus()
		p.Opponents = reply.Players
		if reply.CurrentPlayer == p.ID {
			p.doTurn()
			p.endTurn()
		}
		gameOver = reply.Complete
		time.Sleep(300 * time.Millisecond)
	}
}

func (p *Player) doTurn() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	var randIdx = p.ID
	for randIdx == p.ID {
		randIdx = r.Intn(len(p.Opponents))
	}

	args := CardRequest{Target: p.Opponents[randIdx].ID}
	randIdx = r.Intn(len(p.Hand))
	args.Value = p.Hand[randIdx].Value
	p.callAskForCard(args)
}

func (p *Player) callAskForCard(args CardRequest) {
	reply := CardRequestReply{}
	call("GameServer.AskForCards", &args, &reply)
	p.Hand = append(p.Hand, reply.Cards...)
}

func (p *Player) endTurn() {
	var pairList []Pairs
	var used []int
	//Booo selection sort bad
	for i := 0; i < len(p.Hand); i++ {
		for x := i; x < len(p.Hand); x++ {
			if p.Hand[i].Value == p.Hand[x].Value && notUsed(i, x, used) {
				used = append(used, x)
				used = append(used, i)
				pairList = append(pairList, Pairs{One: p.Hand[i], Two: p.Hand[x]})
			}
		}
	}
	// Idk if this works YA YEET
	for i, v := range used {
		p.Hand = append(p.Hand[:v-i], p.Hand[v+1-i:]...)
	}

	p.callEndTurn(pairList)

}

func (p *Player) callEndTurn(pairs []Pairs) {
	args := PlayPairRequest{Owner: p.ID, Pair: pairs}
	call("GameServer.EndTurn", &args, PlayPairReply{})

	//Might want to update turn here from reply
}

func notUsed(i int, x int, used []int) bool {
	for _, v := range used {
		if v == i || v == x {
			return false
		}
	}
	return true
}

func callGetGameStatus() GameStatusReply {
	reply := GameStatusReply{}
	call("GameServer.AskGameStatus", GameStatusRequest{}, &reply)
	return reply
}

// RPC - Join Game
func callJoinGame() JoinGameReply {
	args := JoinGameArgs{}
	reply := JoinGameReply{}

	call("GameServer.JoinGame", &args, &reply)

	return reply
}

func createPlayer() {
	fmt.Println("CLIENT: successfully created player...")

	// store player state
	me := Player{}

	// keep track of cards locally. if player goes down, server also keeps track

	// join game via RPC
	fmt.Println("CLIENT: joining game...")
	reply := callJoinGame()

	if !reply.Success {
		fmt.Println("Error: Could not join game.")
		return
	}
	// fmt.Println("Joined game")

	me.ID = reply.ID
	// me.Hand = reply.Hand
	// me.Pairs = reply.Pairs

}

func main() {

	createPlayer()

}

// ** copied from mapreduce **
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
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
