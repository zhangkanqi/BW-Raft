package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type State int
const NULL int32 = -1

const (
	Follower State = iota // Follower = 0
	Candidate
	Leader
)

type Op struct {
	option string
	key string
	value string
	id int32
	seq int32
}

type Entry struct {
	term int32
	Command Op
}

type Raft struct {
	mu *sync.Mutex
	me int32
	address string
	members []string
	role State

	currentTerm int32
	votedFor int32
	log []Entry

	commitIndex int32
	lastApplied int32

	nextIndex []int32
	matchIndex []int32

	voteCh chan bool
	appendLogCh chan bool
	heartbeatCh chan bool
	killCh chan bool
	beLeaderCh chan bool

	persist *Persister
}
func getMe(address string) int32 {
	add := strings.Split(address, ".") // 192.168.8.4:5000
	add = strings.Split(add[len(add)-1], ":") // 4:5000
	me, err := strconv.Atoi(add[0])
	if err != nil {
		log.Fatalln(err)
	}
	return int32(me)
}

func MakeRaft(address string, members []string, persist *Persister, mu *sync.Mutex) *Raft {
	raft := &Raft{}
	raft.address = address
	raft.me = getMe(address)
	raft.members = members
	raft.persist = persist
	raft.mu = mu
	n := len(raft.members)
	fmt.Printf("当前节点:%s, rf.me=%d, 所有成员地址：\n", raft.address, raft.me)
	for i := 0; i < n; i++ {
		fmt.Println(raft.members[i])
	}
	return raft
}

func (rf *Raft) init() {
	rf.role = Follower
	rf.currentTerm = 0
	rf.votedFor = NULL
	rf.log = make([]Entry, 1)
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.voteCh = make(chan bool, 1)
	rf.appendLogCh = make(chan bool, 1)
	rf.killCh = make(chan bool, 1)
	rf.heartbeatCh = make(chan bool, 1)
	rf.beLeaderCh = make(chan bool, 1)

	heartbeatTime := time.Duration(20) * time.Millisecond

	go func() {
		for {
			select {
			case <- rf.killCh:
				//rf.persist.Close()
				return
			default:
			}
			electionTime := time.Duration(rand.Intn(20)+500) * time.Millisecond

			switch rf.role {
			case Follower:
				select {
				case <- rf.voteCh:
				case <- rf.appendLogCh:
				case <- time.After(electionTime):
					rf.mu.Lock()
					rf.beCandidate()
					fmt.Printf("%d号节点选举超时，成为candidate，发起新一轮选举，currentTerm=%d\n", rf.me, rf.currentTerm)
					rf.mu.Unlock()
				}
			case Candidate:
				select {
				case <- rf.voteCh:
				case <- rf.appendLogCh:
				case <- time.After(electionTime):
				case <- rf.heartbeatCh: // 新的leader已选出
					rf.mu.Lock()
					rf.beFollower(rf.currentTerm)
					fmt.Printf("新leader已经选出，%d号节点成为Follower，currentTerm=%d\n", rf.me, rf.currentTerm)
					rf.mu.Unlock()
				case <- rf.beLeaderCh:
					rf.mu.Lock()
					rf.beLeader()
					fmt.Printf("%d号节点成为Leader，currentTerm=%d\n", rf.me, rf.currentTerm)
					rf.mu.Unlock()
				}
			case Leader:
				rf.startAppendEntries()
				time.Sleep(heartbeatTime) // 发送心跳包，维持自己的leader地位
			}
		}
	}()

}

func (rf *Raft) startElection() {
	fmt.Printf("############ 开始选举 me:%d term:%d ############\n", rf.me, rf.currentTerm)
	args := &RequestVoteArgs{
		Term:          rf.currentTerm,
		CandidateId:   rf.me,
		LastLogIndex:  rf.getLastLogIndex(),
		LastLogTerm:   rf.getLastLogTerm(),
	}
	var votes int32 = 1 //自己给自己投的一票
	n := len(rf.members)
	for i := 0; i < n; i++ {
		rf.mu.Lock()
		if rf.role != Candidate {
			return
		}
		rf.mu.Unlock()
		if rf.members[i] == rf.address {
			continue
		}
		go func(idx int) {
			reply := rf.sendRequestVote(rf.address, args)
			if reply.Term > rf.currentTerm {
				rf.mu.Lock()
				rf.beFollower(reply.Term)
				rf.mu.Unlock()
				return
			}
			if rf.role != Candidate { // 有其他candidate当选了leader
				return
			}
			if reply.VoteGranted {
				fmt.Printf("%s 获得 %s 的投票\n", rf.address, rf.members[i])
				atomic.AddInt32(&votes, 1)
			}
			if atomic.LoadInt32(&votes) > int32(n/2) {
				fmt.Printf("%s 获得过半的投票，成为leader\n", rf.address)
				rf.mu.Lock()
				rf.beLeader()
				send(rf.voteCh)
				rf.mu.Unlock()
			}
		}(i)
	}
}

func (rf *Raft) sendRequestVote(address string, args *RequestVoteArgs) *RequestVoteReply {
	// RequestVote RPC 中的Client端
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		log.Fatalln(err1)
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			log.Fatalln(err2)
		}
	}()
	client := NewRaftClient(conn)
	reply, err3 := client.RequestVote(context.Background(), args)
	if err3 != nil {
		log.Fatalln(err3)
	}
	return reply
}

func (rf *Raft) RequestVote(ctx context.Context, args *RequestVoteArgs) (*RequestVoteReply, error) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply := &RequestVoteReply{VoteGranted:false}
	reply.Term = rf.currentTerm //用于candidate更新自己的current
	// 发送者：args-client
	// 接收者：rf-server
	// ????根据下面这行代码，已投票的candidate/follower发现自己的term过时后会成为follower（清空votedFor），那之前的投票会被收回吗？
	if rf.currentTerm < args.Term {
		// candidate1 发送RPC到 candidate2，candidate2发现自己的term过时了，candidate2立即变成follower，再判断要不要给candidate1投票
		//？ candidate1 发送RPC到 follower，follower发现的term过时，清空自己的votedFor，再判断要不要给candidate1投票
		rf.beFollower(args.Term)
	}
	if args.Term >= rf.currentTerm && (rf.votedFor == NULL || rf.votedFor == args.CandidateId) &&
		(args.LastLogTerm > rf.getLastLogTerm() ||
			(args.LastLogTerm == rf.getLastLogTerm() && args.LastLogIndex >= rf.getLastLogIndex())) {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		send(rf.voteCh)
	}
	return reply, nil
}

func (rf *Raft) startAppendEntries() {
	fmt.Printf("############ 开始日志追加 me:%d term:%d ############\n", rf.me, rf.currentTerm)
	n := len(rf.members)
	for i := 0; i < n; i++ {
		rf.mu.Lock()
		if rf.role != Leader {
			return
		}
		if rf.members[i] == rf.address {
			continue
		}
		rf.mu.Unlock()
		go func(idx int) {
			appendEntries := rf.log[rf.nextIndex[idx]:] // 刚开始是空的，为了维持自己的leader身份
			entries, _ := json.Marshal(appendEntries) // 转码
			args := &AppendEntriesArgs{
				Term:          rf.currentTerm,
				LeaderId:      rf.me,
				PreLogIndex:   rf.getPrevLogIndex(idx),
				PreLogTerm:    rf.getPrevLogTerm(idx),
				Entries:       entries,
				LeaderCommit:  rf.commitIndex,
			}
			reply := rf.sendAppendEntries(rf.address, args)
			//???为什么要加这个rf.currentTerm != args.Term
			if rf.role != Leader {
				return
			}
			if reply.Term > rf.currentTerm {
				rf.beFollower(reply.Term)
				return
			}
			if reply.Success {
				rf.matchIndex[idx] = args.PreLogIndex + len()
			}
		}(i)
	}

}

func (rf *Raft) sendAppendEntries(address string, args *AppendEntriesArgs) *AppendEntriesReply {

}

func (rf *Raft) AppendEntries(ctx context.Context, args *AppendEntriesArgs) (*AppendEntriesReply, error) {


}

func (rf *Raft) registerServer(address string) {
	//Raft服务的Server端
	server := grpc.NewServer()
	RegisterRaftServer(server, rf)
	lis, err1 := net.Listen("tcp", address)
	if err1 != nil {
		log.Fatalln(err1)
	}
	err2 := server.Serve(lis)
	if err2 != nil {
		log.Fatalln(err2)
	}
}

func send(ch chan bool) {
	select {
	case <- ch: //chan中有内容则先取出
	}
	ch <- true
}

func (rf *Raft) beCandidate() {
	rf.role = Candidate
	rf.currentTerm++
	rf.votedFor = rf.me
	go rf.startElection()
}

func (rf *Raft) beFollower(term int32) {
	rf.role = Follower
	rf.votedFor = NULL
	rf.currentTerm = term
}

func (rf *Raft) beLeader() {
	//if rf.role != Candidate {
	//	return
	//}
	rf.role = Leader
	n :=  len(rf.members)
	rf.nextIndex = make([]int32, n)
	rf.matchIndex = make([]int32, n)
	for i := 0; i < n; i++ {
		rf.nextIndex[i] = rf.getLastLogIndex()+1
	}
}

func (rf *Raft) getLastLogIndex() int32 {
	// 数组下标 0 1 2 3 4
	// len=5, 最新日志索引Index=4，在数组中的下标也为4
	return int32(len(rf.log)-1) // empty, =0(最新日志索引)
}

func (rf *Raft) getLastLogTerm() int32 {
	index := rf.getLastLogIndex()
	if index == 0 { // log is empty
		return -1
	}
	return rf.log[index].term
}

func (rf *Raft) getPrevLogIndex(i int) int32 {
	// 每个peer对应的nextIndex会变，其相应的prevLogIndex也会变
	return rf.nextIndex[i]-1
}

func (rf *Raft) getPrevLogTerm(i int) int32 {
	prevLogIndex := rf.getPrevLogIndex(i)
	if prevLogIndex == 0 { //empty
		return -1
	}
	return rf.log[prevLogIndex].term
}

