package testRaft

import (
	//"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"math/rand"
	"net"
	"sync"
	PERSISTER "../persister"
	RPC "../testRPC"
	"sync/atomic"
	"time"
)

type State int

const (
	Follower  State = iota // value --> 0
	Candidate              // value --> 1
	Leader                 // value --> 2
)
const NULL int32 = -1

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Option string
	Key    string
	Value  string
	Id     int64
	Seq    int64
}

type Log struct {
	Term int32 //  "term when entry was received by leader"
	// Debug 原来是interface{}类型，但是因为json序列化和反序列化时Command类型转化有问题
	Command Op //"command for state machine,"
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int32
}

type Raft struct {
	mu *sync.Mutex // Lock to protect shared access to this peer's state
	// peers     []*labrpc.ClientEnd // RPC end points of all peers
	// persister *Persister          // Object to hold this peer's persisted state
	me int32 // this peer's index into peers[]

	// state a Raft server must maintain.
	state State

	//Persistent state on all servers:(Updated on stable storage before responding to RPCs)
	currentTerm int32 // "latest term server has seen (initialized to 0 increases monotonically)"
	votedFor    int32 // "candidateId that received vote in current term (or null if none)"
	log         []Log // "log entries;(first index is 1)"

	//Volatile state on all servers:
	commitIndex int32 // "index of highest log entry known to be committed (initialized to 0, increases monotonically)"
	lastApplied int32 // "index of highest log entry applied to state machine (initialized to 0, increases monotonically)"

	//Volatile state on leaders：(Reinitialized after election)
	nextIndex  []int32 // "for each server,index of the next log entry to send to that server"
	matchIndex []int32 // "for each server,index of highest log entry known to be replicated on server(initialized to 0, im)"

	//channel
	applyCh chan int  // from Make()
	killCh  chan bool //for Kill()
	//handle rpc
	voteCh      chan bool
	appendLogCh chan bool

	// New
	persist *PERSISTER.Persister
	client  RPC.RaftClient
	address string
	members []string
	delay   int
}

func send(ch chan bool) {
	select {
	case <-ch: //if already set, consume it then resent to avoid block
	default:
	}
	ch <- true
}

func (rf *Raft) getPrevLogIdx(i int) int32 {
	return rf.nextIndex[i] - 1
}

func (rf *Raft) getPrevLogTerm(i int) int32 {
	prevLogIdx := rf.getPrevLogIdx(i)
	if prevLogIdx < 0 {
		return -1
	}
	return rf.log[prevLogIdx].Term
}

func (rf *Raft) getLastLogIdx() int32 {
	return int32(len(rf.log) - 1)
}

func (rf *Raft) getLastLogTerm() int32 {
	idx := rf.getLastLogIdx()
	if idx < 0 {
		return -1
	}
	return rf.log[idx].Term
}

func MakeRaft(add string, mem []string, persist *PERSISTER.Persister,
	mu *sync.Mutex, applyCh chan int, delay int) *Raft {
	raft := &Raft{}
	if len(mem) <= 1 {
		panic("#######Address is less 1, you should set follower's address!######")
	}
	raft.address = add
	raft.persist = persist
	raft.applyCh = applyCh
	raft.mu = mu
	raft.members = make([]string, len(mem))
	raft.delay = delay
	for i := 0; i < len(mem); i++ {
		raft.members[i] = mem[i]
		fmt.Println(raft.members[i])
	}
	raft.init()
	return raft
}

func (rf *Raft) init() {

	rf.state = Follower
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.log = make([]Log, 1) //(first index is 1)

	rf.commitIndex = 0
	rf.lastApplied = 0

	//because gorountne only send the chan to below goroutine,to avoid block, need 1 buffer
	rf.voteCh = make(chan bool, 1)
	rf.appendLogCh = make(chan bool, 1)
	rf.killCh = make(chan bool, 1)

	heartbeatTime := time.Duration(150) * time.Millisecond
	go func() {
		for {
			select {
			case <-rf.killCh:
				return
			default:
			}
			electionTime := time.Duration(rand.Intn(350)+500) * time.Millisecond

			// rf.mu.Lock()
			state := rf.state
			// rf.mu.Unlock()
			switch state {
			case Follower, Candidate:
				select {
				case <-rf.voteCh:
				case <-rf.appendLogCh:
				case <-time.After(electionTime):
					//    rf.mu.Lock()
					fmt.Println("######## time.After(electionTime) #######")
					rf.beCandidate() //becandidate, Reset election timer, then start election
					//    rf.mu.Unlock()
				}
			case Leader:
				//fmt.Println(rf.address, " 当选leader········")
				rf.startAppendLog()
				time.Sleep(heartbeatTime)
			}
		}
	}()

	// Add New

	go rf.RegisterServer(rf.address)

}


func (rf *Raft) startElection() {

	fmt.Println("startElection")

	args := RPC.RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: rf.getLastLogIdx(),
		LastLogTerm:  rf.getLastLogTerm(),
	}
	// TODO.....
	var votes int32 = 1
	for i := 0; i < len(rf.members); i++ {
		if rf.address == rf.members[i] {
			continue
		}
		go func(idx int) {
			//fmt.Println("sendRequestVote to :", rf.members[idx])
			//reply := RPC.RequestVoteReply{Term:9999, VoteGranted: false}
			ret, reply := rf.sendRequestVote(rf.members[idx], &args /* ,&reply */)

			if ret {
				/* rf.mu.Lock()
				   defer rf.mu.Unlock() */
				if reply.Term > rf.currentTerm {
					//fmt.Println( "reply.beFollower ")
					rf.beFollower(reply.Term)
					return
				}
				if rf.state != Candidate || rf.currentTerm != args.Term {
					//fmt.Println("rf.state != Candidate || rf.currentTerm != args.Term")
					return
				}
				if reply.VoteGranted {
					//fmt.Println( "#########reply.VoteGranted ############3")
					atomic.AddInt32(&votes, 1)
				} else {
					//	fmt.Println( "#########reply.VoteGranted false ", reply.Term)
				}
				if atomic.LoadInt32(&votes) > int32(len(rf.members)/2) {
					rf.beLeader()
					//fmt.Println("rf.beLeader()")
					send(rf.voteCh) //after be leader, then notify 'select' goroutine will sending out heartbeats immediately
				}
			}
		}(i)

	}
}

func (rf *Raft) sendRequestVote(address string, args *RPC.RequestVoteArgs) (bool, *RPC.RequestVoteReply) {
	//fmt.
	// Initialize Client
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	rf.client = RPC.NewRaftClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	//var err error
	reply, err := rf.client.RequestVote(ctx, args)
	/* *reply.Term = *r.Term
	*reply.VoteGranted = *r.VoteGranted */
	if err != nil {
		return false, reply
	} else {
		return true, reply

	}
}

func (rf *Raft) RequestVote(ctx context.Context, args *RPC.RequestVoteArgs) (*RPC.RequestVoteReply, error) {
	/*  rf.mu.Lock()
	defer rf.mu.Unlock() */
	fmt.Println("·····1······开始判断是否投票········")
	if args.Term > rf.currentTerm { //all server rule 1 If RPC request or response contains term T > currentTerm:
		rf.beFollower(args.Term) // set currentTerm = T, convert to follower (§5.1)
	}
	reply := &RPC.RequestVoteReply{}
	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	if (args.Term < rf.currentTerm) || (rf.votedFor != NULL && rf.votedFor != args.CandidateId) {
		// Reply false if term < currentTerm (§5.1)  If votedFor is not null and not candidateId,
	} else if args.LastLogTerm < rf.getLastLogTerm() || (args.LastLogTerm == rf.getLastLogTerm() && args.LastLogIndex < rf.getLastLogIdx()) {
		//If the logs have last entries with different terms, then the log with the later term is more up-to-date.
		// If the logs end with the same term, then whichever log is longer is more up-to-date.
		// Reply false if candidate’s log is at least as up-to-date as receiver’s log
	} else {
		//grant vote
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		rf.state = Follower
		// rf.persist()
		send(rf.voteCh) //because If election timeout elapses without receiving granting vote to candidate, so wake up

	}
	// Debug TODO......
	reply.VoteGranted = true
	fmt.Println("······2·····投票判断结束········")
	return reply, nil
}

func (rf *Raft) RegisterServer(address string) {
	// Register Server
	for {

		lis, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		RPC.RegisterRaftServer(s, rf)
		// Register reflection service on gRPC server.
		reflection.Register(s)
		if err := s.Serve(lis); err != nil {
			fmt.Printf("failed to serve: %v \n", err)
		}

	}

}

func (rf *Raft) beFollower(term int32) {
	rf.state = Follower
	rf.votedFor = NULL
	rf.currentTerm = term
	//rf.persist()
}

func (rf *Raft) beLeader() {
	if rf.state != Candidate {
		return
	}
	rf.state = Leader
	//initialize leader data
	rf.nextIndex = make([]int32, len(rf.members))
	rf.matchIndex = make([]int32, len(rf.members))
	for i := 0; i < len(rf.nextIndex); i++ { //(initialized to leader last log index + 1)
		rf.nextIndex[i] = rf.getLastLogIdx() + 1
	}
	fmt.Println(rf.address, "#####become LEADER####", rf.currentTerm)
}

func (rf *Raft) beCandidate() { //Reset election timer are finished in caller
	fmt.Println(rf.address, " become Candidate ", rf.currentTerm)
	rf.state = Candidate
	rf.currentTerm++    //Increment currentTerm
	rf.votedFor = rf.me //vote myself first
	//ask for other's vote
	go rf.startElection() //Send RequestVote RPCs to all other servers
}

func (rf *Raft) startAppendLog() {
	fmt.Println("startAppendLog")

	for i := 0; i < len(rf.members); i++ {
		go func(idx int) {
				fmt.Println(rf.address, " 向 ", rf.members[idx], "append Log")
		}(i)
	}
}