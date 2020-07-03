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

*/

// ** adapted from mapreduce
// func GameServerSock() string {
// 	s := "/var/tmp/824-gs-"
// 	s += strconv.Itoa(os.Getuid())
// 	return s
//

// struct to hold game state info
type GameServer struct {
	Ready   bool
	Players common.Player
	Deck    []common.Card
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
	fmt.Printf("%v\n", gs.Deck)
}

// Create a GameServer
// ** adapted from mapreduce
func MakeGameServer() *GameServer {
	gs := GameServer{}
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
