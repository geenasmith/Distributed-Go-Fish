package raft

import (
	"../labgob"
	"../labrpc"
	"log"
	"sync"
	"time"
)

const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Command   string
	ClientId  int64
	RequestId int64
	Key       string
	Value     string
}

type Result struct {
	Ok        bool
	Err       Err
	Value     string
	ClientId  int64
	RequestId int64
	Leader    bool
	Command   string
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *Raft
	applyCh chan ApplyMsg
	dead    int32 // set by Kill()

	maxraftstate int // snapshot if log grows this big

	// Your definitions here.

	resultCh map[int]chan Result
	data     map[string]string
	ack      map[int64]int64
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	entry := Op{}
	entry.Command = "Get"
	entry.Key = args.Key
	entry.RequestId = args.RequestId
	entry.ClientId = args.ClientId
	result := kv.appendToLog(entry)
	if !result.Ok {
		reply.Leader = false
		return
	}
	reply.Leader = true
	reply.Err = result.Err
	reply.Value = result.Value
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	entry := Op{}
	entry.Command = args.Op
	entry.ClientId = args.ClientId
	entry.RequestId = args.RequestId
	entry.Key = args.Key
	entry.Value = args.Value

	result := kv.appendToLog(entry)
	if !result.Ok {
		reply.Leader = false
		return
	}
	reply.Leader = true
	reply.Err = result.Err

}

func (kv *KVServer) appendToLog(entry Op) Result {
	index, _, leader := kv.rf.Start(entry)

	if !leader {
		return Result{Ok: false}
	}
	kv.mu.Lock()
	if _, ok := kv.resultCh[index]; !ok {
		kv.resultCh[index] = make(chan Result, 1)
	}
	kv.mu.Unlock()

	select {
	case result := <-kv.resultCh[index]:
		if entry.ClientId == result.ClientId && entry.RequestId == result.RequestId {
			return result
		}
		return Result{Ok: false}
	case <-time.After(300 * time.Millisecond):
		return Result{Ok: false}
	}

}

func (kv *KVServer) run() {
	for {
		msg := <-kv.applyCh
		kv.mu.Lock()
		op := msg.Command.(Op)
		result := kv.applyOp(op)
		if ch, ok := kv.resultCh[msg.CommandIndex]; ok {
			select {
			case <-ch:
			default:
			}
		} else {
			kv.resultCh[msg.CommandIndex] = make(chan Result, 1)
		}
		kv.resultCh[msg.CommandIndex] <- result
		kv.mu.Unlock()
	}
}

func (kv *KVServer) applyOp(op Op) Result {
	result := Result{}
	result.Command = op.Command
	result.RequestId = op.RequestId
	result.ClientId = op.ClientId
	result.Leader = true
	result.Ok = true

	if op.Command == "Get" {
		if value, ok := kv.data[op.Key]; ok {
			result.Err = OK
			result.Value = value
		} else {
			result.Err = ErrNoKey
		}
	} else if op.Command == "Put" {
		if !kv.isDup(op) {
			kv.data[op.Key] = op.Value
		}
		result.Err = OK
	} else if op.Command == "Append" {
		if !kv.isDup(op) {
			kv.data[op.Key] += op.Value
		}
		result.Err = OK
	}
	kv.ack[op.ClientId] = op.RequestId
	return result
}

func (kv *KVServer) isDup(op Op) bool {
	lastId, ok := kv.ack[op.ClientId]
	if ok {
		return lastId >= op.RequestId
	} else {
		return false
	}
}

//
// the tester calls Kill() when a KVServer instance won't
// be needed again. for your convenience, we supply
// code to set rf.dead (without needing a lock),
// and a killed() method to test rf.dead in
// long-running loops. you can also add your own
// code to Kill(). you're not required to do anything
// about this, but it may be convenient (for example)
// to suppress debug output from a Kill()ed instance.

func (kv *KVServer) Kill() {
	kv.rf.Kill()
}

//
//
// servers[] contains the ports of the set of
// servers that will cooperate via Raft to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// the k/v server should store snapshots through the underlying Raft
// implementation, which should call persister.SaveStateAndSnapshot() to
// atomically save the Raft state along with the snapshot.
// the k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
// in order to allow Raft to garbage-collect its log. if maxraftstate is -1,
// you don't need to snapshot.
// StartKVServer() must return quickly, so it should start goroutines
// for any long-running work.
//
func StartKVServer(servers []*labrpc.ClientEnd, me int, persister *Persister, maxraftstate int) *KVServer {
	// call labgob.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	labgob.Register(Op{})
	labgob.Register(Result{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	// You may need initialization code here.

	kv.applyCh = make(chan ApplyMsg)
	kv.rf = Make(servers, me, persister, kv.applyCh)

	kv.data = make(map[string]string)
	kv.ack = make(map[int64]int64)
	kv.resultCh = make(map[int]chan Result)
	// You may need initialization code here.
	go kv.run()
	return kv
}
