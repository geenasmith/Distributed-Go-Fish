// **** NOTE: none of this has been tested. Just copy/pasted and changed slightly from mapreduce

package main

import "encoding/json"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "log"
import "./common"

// ** adapted from mapreduce
// func GameServerSock() string {
// 	s := "/var/tmp/824-gs-"
// 	s += strconv.Itoa(os.Getuid())
// 	return s
// }

type Card struct {
	Value int
	Suite string
}

// struct to hold game state info
type GameServer struct {
	Ready bool
	Deck  []Card
}

type Pairs struct {
	one Card
	two Card
}

type Player struct {
	Hand []Card
	Pars []Pairs
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
	var values []Card
	file, err := os.Open("standard52.json")
	if err != nil {
		log.Fatalf("Can opend card file")
	}
	dec := json.NewDecoder(file)
	for {
		var item Card
		if err := dec.Decode(&item); err != nil {
			if err != io.EOF {
				fmt.Printf("%v\n", err)
			}
			break
		}
		values = append(values, item)
	}
	_ = file.Close()
	fmt.Printf("%v", values)
}

// Create a GameServer
// ** adapted from mapreduce
func MakeGameServer() *GameServer {
	gs := GameServer{}
	gs.loadCards()
	gs.server()
	return &gs
}

func main(){

	// ** adapted from mapreduce
	gs := MakeGameServer()

	fmt.Println("successfully created server...")


}
