
package main

import "net/rpc"
import "fmt"
import "log"
import "./common"

// 
func callJoinGame() common.JoinGameReply {
	args := common.JoinGameArgs{}
	reply := common.JoinGameReply{}

	call("GameServer.JoinGame", &args, &reply)

	return reply
}

func Player(){
	fmt.Println("successfully created player...")

	// keep track of cards locally. if player goes down, server also keeps track

	// join game via RPC
	fmt.Println("joining game...")
	reply := callJoinGame()

	if reply.Success {
		fmt.Println("\tcallJoinGame working")
	} else {
		fmt.Println("\tcallJoinGame NOT working")
	}

}


func main(){

	Player()

}

// ** copied from mapreduce **
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := common.GameServerSock()
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
