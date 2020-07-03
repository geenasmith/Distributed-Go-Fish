package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"net/rpc"
	"net/http"
	"log"
	"./common"
	"time"
)

// struct to hold game state info
type GameServer struct {
	Ready             bool
	GameOver          bool
	Winner            int
	Players           []common.Player // holds ID, hand, and pairs
	Deck              []common.Card // hold the cards that are still in the deck
	CurrentTurnPlayer int
	CurrentTurn       int
	PlayerCount       int
}

// RPC - Join Game
func (gs *GameServer) JoinGame(args *common.JoinGameArgs, reply *common.JoinGameReply) error {
	reply.Success = true

	// if a game does not exist, start one
	// if PlayerCount == 7, return false (too many players)
	// else add new player
	//     return 7 cards
	//     assign player ID (index 0,1,2,...)


	return nil
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

// helper function to test. Delete later
func (gs *GameServer) initPlayers(x int) {
	gs.PlayerCount = x
	for i := 0; i < gs.PlayerCount; i++ {
		gs.Players = append(gs.Players, common.Player{})
	}
}

func (gs *GameServer) AskIfOver(gameStatus *common.GameStatusReply) {
	gameStatus.Complete = gs.GameOver
	gameStatus.CurrentPlayer = gs.CurrentTurnPlayer
	gameStatus.Turn = gs.CurrentTurn
	var scores []int
	for _, v := range gs.Players {
		scores = append(scores, len(v.Pairs))
	}
	gameStatus.Scores = scores
	if gs.GameOver {
		gameStatus.Winner = gs.Winner
	} else {
		gameStatus.Winner = -1
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
	gs.checkGameOver()
}

func (gs *GameServer) EndTurn(ask *common.PlayPairRequest, reply *common.PlayPairReply) {
	if len(ask.Pair) != 0 {
		for _, v := range ask.Pair {
			gs.Players[ask.Owner].Pairs = append(gs.Players[ask.Owner].Pairs, v)
		}
	}
	gs.CurrentTurnPlayer++
	if gs.CurrentTurnPlayer > gs.PlayerCount {
		gs.CurrentTurn++
	}
}

func (gs *GameServer) goFish() common.Card {
	var fish = gs.Deck[0]
	gs.Deck = gs.Deck[1:]
	return fish
}

func (gs *GameServer) checkGameOver() {
	var deckEmpty = false
	var playerEmpty = true
	if len(gs.Deck) == 0 {
		deckEmpty = true
	}
	for _, v := range gs.Players {
		if len(v.Hand) != 0 {
			playerEmpty = false
		}
	}
	gs.GameOver = playerEmpty && deckEmpty
}

// Create a GameServer
func MakeGameServer() *GameServer {
	gs := GameServer{}
	gs.CurrentTurn = 0
	gs.CurrentTurnPlayer = 0
	gs.GameOver = false
	var x = 3
	gs.initPlayers(x)
	gs.loadCards()
	gs.dealCards()
	gs.server()
	return &gs
}

func main() {

	gs := MakeGameServer()

	for gs.Done() == false {
		time.Sleep(time.Second)
	}

	fmt.Println("successfully created server...")

}
