package main

import (
	"sync"
	"testing"
	"time"
)
import "../raft"

type GameStateArgs struct {
	Key     string
	Payload string
}
type GameStateReply struct {
	Payload string
	Ok      bool
}

type RaftBroker struct {
	mu sync.Mutex
	ck *raft.Clerk
}

func (rb *RaftBroker) PutGameState(args *GameStateArgs, reply *GameStateReply) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.ck.Put(args.Key, args.Payload)
	reply.Ok = true
	return
}
func (rb *RaftBroker) GetGameState(args *GameStateArgs, reply *GameStateReply) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	reply.Payload = rb.ck.Get(args.Key)
	return
}

func main() {
	rb := RaftBroker{}
	const nservers = 5
	var t *testing.T
	cfg := raft.Make_config(t, nservers, false, -1)
	defer cfg.Cleanup()
	rb.ck = cfg.MakeClient(cfg.All())
	for true {
		time.Sleep(3 * time.Second)
	}
}
