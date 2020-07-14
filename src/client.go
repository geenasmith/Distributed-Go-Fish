package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
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
	used  bool
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
	Hand  []Card
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
		if reply.Complete {
			fmt.Printf("Game Done Client\n")
			return
		}
		p.Opponents = reply.Players
		p.Hand = reply.Players[p.ID].Hand
		p.Pairs = reply.Players[p.ID].Pairs
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
	if len(p.Opponents) > 1 {
		fmt.Printf("%v My hand %d\n", p.Hand, p.ID)
		fmt.Printf("%v My paris %d\n\n\n\n", p.Pairs, p.ID)
		for randIdx == p.ID {
			randIdx = r.Intn(len(p.Opponents))
		}
		args := CardRequest{Target: p.Opponents[randIdx].ID}
		if len(p.Hand) > 0 {
			randIdx = r.Intn(len(p.Hand))
			args.Value = p.Hand[randIdx].Value
		} else {
			args.Value = "-1"
		}

		p.callAskForCard(args)
	} else {
		time.Sleep(300 * time.Millisecond)
	}
}

func (p *Player) callAskForCard(args CardRequest) {
	reply := CardRequestReply{}
	call("GameServer.AskForCards", &args, &reply)
	if len(reply.Cards) > 1 || reply.Cards[0].Value != "-1" {
		p.Hand = append(p.Hand, reply.Cards...)
	}
}

func (p *Player) endTurn() {
	var pairList []Pairs
	//Booo selection sort bad
	for i := 0; i < len(p.Hand); i++ {
		for x := i; x < len(p.Hand); x++ {
			if p.Hand[i].Value == p.Hand[x].Value && !p.Hand[x].used && !p.Hand[i].used && i != x {
				p.Hand[i].used = true
				p.Hand[x].used = true
				pairList = append(pairList, Pairs{One: p.Hand[i], Two: p.Hand[x]})
			}
		}

	}

	// Idk if this works YA YEET
	var newList []Card
	for _, v := range p.Hand {
		if v.used != true {
			newList = append(newList, v)
		}
	}
	p.Hand = newList
	p.callEndTurn(pairList)

}

func (p *Player) callEndTurn(pairs []Pairs) {
	args := PlayPairRequest{Owner: p.ID, Pair: pairs, Hand: p.Hand}
	call("GameServer.EndTurn", &args, &PlayPairReply{})

	//Might want to update turn here from reply
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

func createPlayer() Player {
	fmt.Println("CLIENT: successfully created player...")

	// store player state
	me := Player{}

	// keep track of cards locally. if player goes down, server also keeps track

	// join game via RPC
	fmt.Println("CLIENT: joining game...")
	reply := callJoinGame()

	if !reply.Success {
		fmt.Println("Error: Could not join game.")
		os.Exit(1)
	}
	// fmt.Println("Joined game")

	me.ID = reply.ID
	//me.Hand = reply.Hand
	// me.Pairs = reply.Pairs
	return me

}

func runClient() {

	player := createPlayer()
	go player.PlayGoFish()

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
