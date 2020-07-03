// **** NOTE: none of this has been tested. Just copy/pasted and changed slightly from mapreduce

package main

import (
	"encoding/json"
	"fmt"
	"io"
)
import "net"
import "os"
import "net/rpc"
import "net/http"
import "log"
import "./common"

/*
Starting Hand 7 Cards
Max Players 7
Pass All valued cards on ask

*/

// ** adapted from mapreduce
// func GameServerSock() string {
// 	s := "/var/tmp/824-gs-"
// 	s += strconv.Itoa(os.Getuid())
// 	return s
//

// struct to hold game state info
type GameServer struct {
	Ready             bool
	Players           []common.Player
	Deck              []common.Card
	CurrentTurnPlayer int
	CurrentTurn       int
	PlayerCount       int
}

// ** adapted from mapreduce
func (gs *GameServer) Done() bool {
	ret := false

	return ret
}

// ** adapted from mapreduce
func (gs *GameServer) server() {
	rpc.Register(gs)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := common.GameServerSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (gs *GameServer) loadCards() {
	fmt.Printf("loading cards\n")
	var values []common.Card
	file, err := os.Open("standard52.json")
	if err != nil {
		log.Fatalf("Can opend card file\n")
	}
	dec := json.NewDecoder(file)
	if err := dec.Decode(&values); err != nil {
		if err != io.EOF {
			fmt.Printf("%v\n", err)
		}
	}

	_ = file.Close()
	gs.Deck = values
	fmt.Printf("%v Cards\n", values)
}

func (gs *GameServer) dealCards() {
	common.Shuffle(gs.Deck)
	for i := 0; i < gs.PlayerCount; i++ {
		gs.Players[i].Hand = gs.Deck[0:7]
		gs.Deck = gs.Deck[7:]
	}
	for k, v := range gs.Players {
		fmt.Printf("\n\n%v Player %d\n\n", v, k)
	}
}

func (gs *GameServer) initPlayers(x int) {
	gs.PlayerCount = x
	for i := 0; i < gs.PlayerCount; i++ {
		gs.Players = append(gs.Players, common.Player{})
	}

}

func (gs *GameServer) AskForCards(ask *common.CardRequest, reply *common.CardRequestReply) {
	reply.GoFish = true
	var toRemove []int
	var cardPool = gs.Players[ask.Target].Hand
	for k, v := range cardPool {
		if v.Value == ask.Value {
			reply.GoFish = false
			reply.Cards = append(reply.Cards, v)
			toRemove = append(toRemove, k)
		}
	}
	if reply.GoFish {
		reply.Cards = append(reply.Cards, gs.goFish())
	} else {
		for i, v := range toRemove {
			gs.Players[ask.Target].Hand = append(gs.Players[ask.Target].Hand[:v-i], gs.Players[ask.Target].Hand[v+1-i:]...)
		}
	}
}

func (gs *GameServer) goFish() common.Card {
	var fish = gs.Deck[0]
	gs.Deck = gs.Deck[1:]
	return fish
}

// Create a GameServer
// ** adapted from mapreduce
func MakeGameServer() *GameServer {
	gs := GameServer{}
	var x = 3
	gs.initPlayers(x)
	gs.loadCards()
	gs.dealCards()
	gs.server()
	return &gs
}

func main() {

	// ** adapted from mapreduce
	_ = MakeGameServer()

	fmt.Println("successfully created server...")

}
