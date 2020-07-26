package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"testing"
	"time"
)
import "../raft"
import "../helpers"

type RaftBroker struct {
	mu sync.Mutex
	ck *raft.Clerk
}

func (rb *RaftBroker) PutGameState(args *helpers.GameStateArgs, reply *helpers.GameStateReply) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.ck.Put(args.Key, args.Payload)
	reply.Ok = true
	return nil
}
func (rb *RaftBroker) GetGameState(args *helpers.GameStateArgs, reply *helpers.GameStateReply) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	reply.Payload = rb.ck.Get(args.Key)
	reply.Ok = true
	return nil
}

func (rb *RaftBroker) server() {
	rpc.Register(rb)
	rpc.HandleHTTP()
	sockname := helpers.RaftServerSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func main() {
	rb := RaftBroker{}
	rb.server()
	const nservers = 5
	var t *testing.T
	cfg := raft.Make_config(t, nservers, false, -1)
	defer cfg.Cleanup()
	rb.ck = cfg.MakeClient(cfg.All())
	for true {
		time.Sleep(3 * time.Second)
	}
}
