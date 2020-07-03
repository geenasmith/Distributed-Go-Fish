// **** NOTE: none of this has been tested. Just copy/pasted and changed slightly from mapreduce

package main

import "fmt"
import "net"
import "os"
import "net/rpc"
import "net/http"
// import "strconv"
import "log"
import "time"
import "./common"

// ** adapted from mapreduce
// func GameServerSock() string {
// 	s := "/var/tmp/824-gs-"
// 	s += strconv.Itoa(os.Getuid())
// 	return s
// }

// struct to hold game state info
type GameServer struct {

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

// Create a GameServer
// ** adapted from mapreduce
func MakeGameServer() *GameServer {
	gs := GameServer{}
	gs.server()
	return &gs
}

func main(){

	// ** adapted from mapreduce
	gs := MakeGameServer()

	fmt.Println("successfully created server...")

	for gs.Done() == false {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

}
