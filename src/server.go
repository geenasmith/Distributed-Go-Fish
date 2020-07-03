// **** NOTE: none of this has been tested. Just copy/pasted and changed slightly from mapreduce


import "net"
import "os"
import "net/rpc"
import "net/http"

// struct to hold game state info
type GameServer struct {

}

// ** adapted from mapreduce
func (gs *GameServer) Done() bool {
	ret := true

	return ret
}

// ** adapted from mapreduce
func (gs *GameServer) server() {
	rpc.Register(gs)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := gameServerSock()
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
	gs := mr.MakeGameServer()

	for gs.Done() == false {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

}
