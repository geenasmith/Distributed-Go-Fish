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
	"time"
	"sync"

)

// struct to hold game state info
type GameServer struct {
	Mu                sync.Mutex
	Ready             bool
	GameOver          bool
	Winner            int
	Players           []Player // holds ID, hand, and pairs
	Deck              []Card   // hold the cards that are still in the deck
	CurrentTurnPlayer int
	CurrentTurn       int
	PlayerCount       int
	GameInitialized   bool
}

// RPC - Join Game
func (gs *GameServer) JoinGame(args *JoinGameArgs, reply *JoinGameReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	// no more than 7 players
	if gs.PlayerCount == 7 {
		reply.Success = false
		return nil
	}

	reply.Success = true

	// if game doesn't exist, start one
	if !gs.GameInitialized {
		gs.GameInitialized = true
		gs.CurrentTurnPlayer = 0
	}

	reply.ID = gs.PlayerCount

	// append player to game state
	gs.Players = append(gs.Players, Player{ID: gs.PlayerCount})
	gs.PlayerCount++

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
	sockname := GameServerSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (gs *GameServer) loadCards() {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	fmt.Printf("SERVER: loading cards\n")
	var values []Card
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
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	Shuffle(gs.Deck)

	for i := 0; i < gs.PlayerCount; i++ {
		gs.Players[i].Hand = gs.Deck[0:7]
		gs.Deck = gs.Deck[7:]
	}

	fmt.Println("SERVER: Players")
	for k, v := range gs.Players {
		fmt.Printf("\n%v Player %d\n\n", v, k)
	}
}

func (gs *GameServer) AskGameStatus(ask *GameStatusRequest, gameStatus *GameStatusReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	gameStatus.Complete = gs.GameOver
	gameStatus.CurrentPlayer = gs.CurrentTurnPlayer
	gameStatus.Turn = gs.CurrentTurn
	gameStatus.Players = gs.Players
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

	return nil
}

func (gs *GameServer) AskForCards(ask *CardRequest, reply *CardRequestReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

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

	return nil
}

func (gs *GameServer) EndTurn(ask *PlayPairRequest, reply *PlayPairReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	if ask.Pair != nil && len(ask.Pair) != 0 {
		for _, v := range ask.Pair {
			gs.Players[ask.Owner].Pairs = append(gs.Players[ask.Owner].Pairs, v)
		}
	}
	gs.CurrentTurnPlayer++
	if gs.CurrentTurnPlayer > gs.PlayerCount {
		gs.CurrentTurn++
	}

	return nil
}

func (gs *GameServer) goFish() Card {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	var fish = gs.Deck[0]
	gs.Deck = gs.Deck[1:]
	return fish
}

func (gs *GameServer) checkGameOver() {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

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
	gs.GameInitialized = false
	gs.Ready = false

	gs.server()

	// wait 5 seconds for players to join
	// break after time or when 7 players join
	for start := time.Now(); time.Since(start) < 30*time.Second; {
		if gs.PlayerCount == 7 {
			break
		}
	}

	fmt.Printf("\nSERVER: total %d players\n", gs.PlayerCount)

	// not enough players or no one joined
	if gs.PlayerCount < 2 || !gs.GameInitialized {
		fmt.Println("SERVER: Game error")
		return &gs
	}

	gs.loadCards()
	gs.dealCards()

	return &gs
}

func main() {

	gs := MakeGameServer()

	for gs.Done() == false { // modify this later for checking if game done
		time.Sleep(time.Second)
	}

	fmt.Println("SERVER: successfully created server...")

}
