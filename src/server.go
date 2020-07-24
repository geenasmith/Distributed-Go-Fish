package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// RPC for clients (players) to join the game
func (gs *GameServer) JoinGame(args *JoinGameArgs, reply *JoinGameReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	// game can have no more than 7 players
	if gs.PlayerCount == 7 {
		reply.Success = false
		return nil
	}

	reply.Success = true

	// if game doesn't exist, start one
	if !gs.GameInitialized {
		gs.GameInitialized = true
	}

	reply.ID = gs.PlayerCount

	// append player to game state
	gs.Players = append(gs.Players, Player{ID: gs.PlayerCount})
	gs.PlayerCount = len(gs.Players)

	return nil
}

// Load card objects from a JSON file and populate the deck
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
}

// Assign each player 7 cards from the deck
func (gs *GameServer) dealCards() {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	Shuffle(gs.Deck)

	for i := 0; i < gs.PlayerCount; i++ {
		gs.Players[i].Hand = gs.Deck[0:7]
		gs.Deck = gs.Deck[7:]
	}

	fmt.Println("SERVER: Dealing Cards")
}

// RPC for clients (players) to ask the status of the game
func (gs *GameServer) AskGameStatus(ask *GameStatusRequest, gameStatus *GameStatusReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	// information returned to the player includes:
	// - if game is over
	// - current player whose turn it is
	// - the current turn
	// - all players still in the game
	// - all player's scores
	// - winner of the game (if there is one)

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

// RPC for clients (players) to ask for specific cards from a specific player
func (gs *GameServer) AskForCards(ask *CardRequest, reply *CardRequestReply) error {
	gs.Mu.Lock()

	reply.GoFish = true

	// Loop through target player's hand and find matching cards
	var toRemove []int
	var cardPool = gs.Players[ask.Target].Hand
	for k, v := range cardPool {
		if v.Value == ask.Value {
			reply.GoFish = false
			reply.Cards = append(reply.Cards, v)
			toRemove = append(toRemove, k)
		}
	}

	if reply.GoFish { // No card found
		reply.Cards = append(reply.Cards, gs.goFish())
	} else { // Target player has 1 or more matching cards
		sort.Ints(toRemove)
		for i, v := range toRemove {
			gs.Players[ask.Target].Hand = append(gs.Players[ask.Target].Hand[:v-i], gs.Players[ask.Target].Hand[v+1-i:]...)
		}
	}

	gs.Mu.Unlock()

	gs.checkGameOver()

	return nil
}

// RPC for clients (players) to end their turn by playing their pairs and updating the game state
func (gs *GameServer) EndTurn(ask *PlayPairRequest, reply *PlayPairReply) error {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()

	// Update the player's matching pairs
	if ask.Pair != nil && len(ask.Pair) != 0 {
		gs.Players[ask.Owner].Pairs = append(gs.Players[ask.Owner].Pairs, ask.Pair...)
	}

	// Update the player's hand
	gs.Players[ask.Owner].Hand = ask.Hand


	// Determine next player
	gs.CurrentTurnPlayer++
	if gs.CurrentTurnPlayer >= gs.PlayerCount {
		gs.CurrentTurnPlayer = 0
	}
	ask.Pair = nil
	return nil
}

// Go-fish action which draws 1 card
func (gs *GameServer) goFish() Card {

	// Deck empty
	if len(gs.Deck) == 0 {
		return Card{Value: "-1"}
	}

	var fish = gs.Deck[0]
	gs.Deck = gs.Deck[1:]

	return fish
}

// Check if the game is finished, indicated by an empty deck and all players having an empty hand
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
			break
		}
	}

	gs.GameOver = playerEmpty && deckEmpty

	// find the winner, indicated by the greatest number of pairs
	gs.Winner = 0
	for _, k := range gs.Players {
		if len(gs.Players[gs.Winner].Pairs) < len(k.Pairs) {
			gs.Winner = k.ID
		}
	}
}

func (gs *GameServer) server() {
	rpc.Register(gs)
	rpc.HandleHTTP()
	sockname := GameServerSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// Create a GameServer
func MakeGameServer() *GameServer {

	gs := GameServer{}

	gs.Mu.Lock()

	gs.CurrentTurn = 0
	gs.CurrentTurnPlayer = -1
	gs.GameOver = false
	gs.GameInitialized = false
	gs.Ready = false

	gs.server()

	gs.Mu.Unlock()

	// create 2 players
	//go runClient()
	//go runClient()

	time.Sleep(3 * time.Second)

	fmt.Printf("\nSERVER: total %d players\n", gs.PlayerCount)

	// not enough players or no one joined
	if gs.PlayerCount < 2 {
		fmt.Println("SERVER: Game error not enough players")
		return &gs
	}

	gs.loadCards()
	gs.dealCards()

	gs.CurrentTurnPlayer = 0

	return &gs
}

func main() {

	gs := MakeGameServer()
	gs.CurrentTurn = 0
	gs.saveGameState()
	gs.CurrentTurn = 15
	gs.getGameState()
	fmt.Printf("%v", gs.CurrentTurn)
	for !gs.GameOver {
		time.Sleep(3 * time.Second)
	}
	fmt.Printf("SERVER: Game Over\n")

	fmt.Printf("SERVER: Player %d won with %d pairs\n", gs.Winner, len(gs.Players[gs.Winner].Pairs))

}

func (gs *GameServer) saveGameState() {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	args := GameStateArgs{}
	reply := GameStateReply{}
	args.Key = string(gs.ServerId)
	js, _ := json.Marshal(gs)
	args.Payload = string(js)
	ok := call("RaftBroker.PutGameState", &args, &reply)
	if !ok || !reply.OK {
		fmt.Printf("Put Game state failed\n")
	}
}

func (gs *GameServer) getGameState() {
	gs.Mu.Lock()
	defer gs.Mu.Unlock()
	args := GameStateArgs{}
	reply := GameStateReply{}
	args.Key = string(gs.ServerId)
	ok := call("RaftBroker.PutGameState", &args, &reply)
	if !ok || !reply.OK {
		fmt.Printf("Put Game state failed\n")
	}
	gs.reconcileState(reply.Payload)

}

func (gs *GameServer) reconcileState(payload string) {
	var gsSaved GameServer
	err := json.Unmarshal([]byte(payload), &gsSaved)
	if err != nil {
		fmt.Printf("Unmarshall of game state failed")
	}
	if gsSaved.ServerId != gs.ServerId {
		fmt.Printf("Wrong game state retreived")
	} else {
		gs.Winner = gsSaved.Winner
		gs.Players = gsSaved.Players
		gs.GameOver = gsSaved.GameOver
		gs.CurrentTurnPlayer = gsSaved.CurrentTurnPlayer
		gs.CurrentTurn = gsSaved.CurrentTurn
		gs.Deck = gsSaved.Deck
		gs.Ready = gsSaved.Ready
		gs.GameInitialized = gsSaved.GameInitialized
	}

}
