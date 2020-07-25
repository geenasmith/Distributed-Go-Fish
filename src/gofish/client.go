package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)
import "../helpers"



type Player struct {
	ID        int
	Hand      []helpers.Card
	Pairs     []helpers.Pairs
	Opponents []Player
}

type GameStatusReply struct {
	Complete      bool
	Turn          int
	CurrentPlayer int
	Winner        int
	Scores        []int
	Players       []Player
}



// Main player functionality
func (p *Player) PlayGoFish() {

	var gameOver = false
	for !gameOver {
		var reply = callGetGameStatus()

		if reply.Complete {
			fmt.Printf("CLIENT: Game Over\n")
			return
		}

		// update the player's game state information
		p.Opponents = reply.Players
		p.Hand = reply.Players[p.ID].Hand
		p.Pairs = reply.Players[p.ID].Pairs

		// Currently this player's turn
		if reply.CurrentPlayer == p.ID {
			p.doTurn()
			p.endTurn()
		}

		gameOver = reply.Complete

		time.Sleep(300 * time.Millisecond)
	}
}

// Determine the card value and opponent to ask. Send RPC to the server
func (p *Player) doTurn() {

	r := rand.New(rand.NewSource(time.Now().Unix()))

	// randomly choose an opponent to ask
	var randIdx = p.ID
	if len(p.Opponents) > 1 {
		fmt.Printf("Player %d my hand: %v\n", p.ID, p.Hand)
		fmt.Printf("Player %d my pairs: %v\n\n", p.ID, p.Pairs)

		for randIdx == p.ID {
			randIdx = r.Intn(len(p.Opponents))
		}

		args := helpers.CardRequest{Target: p.Opponents[randIdx].ID}

		// randomly choose a card value from your hand to ask for
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

// Send RPC to the server asking for a card value from an opponent
func (p *Player) callAskForCard(args helpers.CardRequest) {
	reply := helpers.CardRequestReply{}
	helpers.CallGS("GameServer.AskForCards", &args, &reply)

	// append any returned cards to your hand
	if len(reply.Cards) > 1 || reply.Cards[0].Value != "-1" {
		p.Hand = append(p.Hand, reply.Cards...)
	}
}

// Find matching pairs in your hand, remove them and send to the server
func (p *Player) endTurn() {

	// find matching pairs
	var pairList []helpers.Pairs
	for i := 0; i < len(p.Hand); i++ {
		for x := i; x < len(p.Hand); x++ {
			if p.Hand[i].Value == p.Hand[x].Value && !p.Hand[x].Used && !p.Hand[i].Used && i != x {
				p.Hand[i].Used = true
				p.Hand[x].Used = true
				pairList = append(pairList, helpers.Pairs{One: p.Hand[i], Two: p.Hand[x]})
			}
		}

	}

	// remove pairs from player's hand
	var newList []helpers.Card
	for _, v := range p.Hand {
		if v.Used != true {
			newList = append(newList, v)
		}
	}

	p.Hand = newList

	// send RPC to server notifying of removed pairs and to let the next player go
	p.callEndTurn(pairList)

}

// Send RPC to the server notifying of removed pairs and to let the next player go
func (p *Player) callEndTurn(pairs []helpers.Pairs) {
	args := helpers.PlayPairRequest{Owner: p.ID, Pair: pairs, Hand: p.Hand}
	helpers.CallGS("GameServer.EndTurn", &args, &helpers.PlayPairReply{})

	//Might want to update turn here from reply
}

// Send RPC to the server to get the current game state
func callGetGameStatus() GameStatusReply {
	reply := GameStatusReply{}
	helpers.CallGS("GameServer.AskGameStatus", helpers.GameStatusRequest{}, &reply)
	return reply
}

// Send RPC to the server asking to join game
func callJoinGame() helpers.JoinGameReply {
	args := helpers.JoinGameArgs{}
	reply := helpers.JoinGameReply{}

	helpers.CallGS("GameServer.JoinGame", &args, &reply)

	return reply
}

// Create the new player object
func createPlayer() Player {
	fmt.Println("CLIENT: successfully created player...")

	// store player state
	me := Player{}


	// join game via RPC
	fmt.Println("CLIENT: joining game...")
	reply := callJoinGame()

	if !reply.Success {
		fmt.Println("Error: Could not join game.")
		os.Exit(1)
	}

	me.ID = reply.ID

	return me

}

func runClient() {

	player := createPlayer()
	go player.PlayGoFish()

}
