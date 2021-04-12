package Raft

import (
	RPC "../RPC"
	PERSISTER "../persist"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type State int
type IntSlice []int32
const NULL int32 = -1

const (
	Follower State = iota // Follower = 0
	Candidate
	Leader
)

type Op struct {
	Option string
	Key    string
	Value  string
	Id     int32
	Seq    int32
}

type Entry struct {
	term int32
	Command Op
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type Raft struct {
	mu *sync.Mutex
	me int32
	address string
	members []string //其他成员，包括自己
	role State

	currentTerm int32
	votedFor int32
	votes int32
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
	applyCh chan ApplyMsg

	Persist *PERSISTER.Persister
}
func getMe(address string) int32 {
	add := strings.Split(address, ".") // 192.168.8.4:5000
	add = strings.Split(add[len(add)-1], ":") // 4:5000
	me, err := strconv.Atoi(add[0])
	if err != nil {
		fmt.Println(err)
	}
	return int32(me)
}

// 新指令的index，term，isLeader
func (rf *Raft) Start(command interface{}) (int32, int32, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	var index int32 = -1
	term := rf.currentTerm
	isLeader := rf.role == Leader
	if isLeader {
		index = rf.getLastLogIndex() + 1
		newEntry := Entry{
			term:    rf.currentTerm,
			Command: command.(Op), //？
		}
		rf.log = append(rf.log, newEntry)
		rf.startAppendEntries()
	}
	fmt.Printf("新日志的Index：%d，term：%d，内容：%s\n", index, term, command)
	return index, term, isLeader
}

func (rf *Raft) registerServer(address string) {
	//Raft服务的Server端
	for {
		server := grpc.NewServer()
		RPC.RegisterRaftServer(server, rf)
		lis, err1 := net.Listen("tcp", address)
		if err1 != nil {
			fmt.Println(err1)
		}
		err2 := server.Serve(lis)
		if err2 != nil {
			fmt.Println(err2)
		}
		fmt.Println("····················注册服务器成功······················")
	}
}

func MakeRaft(address string, members []string, persist *PERSISTER.Persister, mu *sync.Mutex, ) *Raft {
	raft := &Raft{}
	raft.address = address
	raft.me = getMe(address)
	raft.members = members
	raft.Persist = persist
	raft.mu = mu
	n := len(raft.members)
	fmt.Printf("当前节点:%s, rf.me=%d, 所有成员地址：\n", raft.address, raft.me)
	for i := 0; i < n; i++ {
		fmt.Println(raft.members[i])
	}
	raft.init()
	return raft
}

func (rf *Raft) init() {
	rf.role = Follower
	rf.currentTerm = 0
	rf.votedFor = NULL
	rf.log = make([]Entry, 1) // 日志索引从1开始
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.voteCh = make(chan bool, 1)
	rf.appendLogCh = make(chan bool, 1)
	rf.killCh = make(chan bool, 1)
	rf.heartbeatCh = make(chan bool, 1)
	rf.beLeaderCh = make(chan bool, 1)

	heartbeatTime := time.Duration(150) * time.Millisecond

	go func() {
		for {
			select {
			case <- rf.killCh:
				//rf.persist.Close()
				return
			default:
			}
			electionTime := time.Duration(rand.Intn(350)+500) * time.Millisecond
			//electionTime := time.Second
			switch rf.role {
			case Follower:
				select {
				case <- rf.voteCh:
				case <- rf.appendLogCh:
				case <- time.After(electionTime):
					fmt.Printf("%d号节点没有收到心跳包，成为candidate，发起新一轮选举，旧的currentTerm=%d\n", rf.me, rf.currentTerm)
					rf.beCandidate()
				}
			case Candidate:
				select {
				case <- rf.voteCh:
				case <- rf.appendLogCh:
				case <- time.After(electionTime):
					fmt.Printf("%d号节点选举超时，成为candidate，发起新一轮选举，旧的currentTerm=%d\n", rf.me, rf.currentTerm)
					rf.beCandidate()
					//case <- rf.heartbeatCh: // 新的leader已选出
					//rf.beFollower(rf.currentTerm)
					//fmt.Printf("新leader已经选出，%d号节点成为Follower，currentTerm=%d\n", rf.me, rf.currentTerm)
					//case <- rf.beLeaderCh:
					//fmt.Printf("%d号节点成为Leader，currentTerm=%d\n", rf.me, rf.currentTerm)
				}
			case Leader:
				rf.startAppendEntries()
				time.Sleep(heartbeatTime) // 发送心跳包，维持自己的leader地位
			}
		}
	}()

	go rf.registerServer(rf.address)

}

func (rf *Raft) startElection() {
	fmt.Printf("############ 开始选举 me:%d term:%d ############\n", rf.me, rf.currentTerm)
	args := &RPC.RequestVoteArgs{
		Term:          rf.currentTerm,
		CandidateId:   rf.me,
		LastLogIndex:  rf.getLastLogIndex(),
		LastLogTerm:   rf.getLastLogTerm(),
	}
	var votes int32 = 1 //自己给自己投的一票
	n := len(rf.members)
	for i := 0; i < n; i++ {
		//rf.mu.Lock()
		if rf.role != Candidate {
			fmt.Println("Candidate 角色变更")
			return
		}
		//rf.mu.Unlock()
		if rf.members[i] == rf.address {
			continue
		}
		go func(idx int) {
			//rf.mu.Lock()
			if rf.role != Candidate {
				return
			}
			//rf.mu.Unlock()
			fmt.Printf("%s 向 %s 发起send RequestVote\n", rf.address, rf.members[idx])
			//原本把rf.members[idx]写成了rf.address
			ret, reply := rf.sendRequestVote(rf.members[idx], args) //一定要有ret
			if ret {
				fmt.Println("RequestVote成功返回结果")
				if reply.Term > rf.currentTerm { // 此Candidate的term过时
					fmt.Println(rf.me, " 的term过期，转成follower")
					rf.beFollower(reply.Term)
					return
				}
				if rf.role != Candidate || rf.currentTerm != args.Term{ // 有其他candidate当选了leader
					return
				}
				if reply.VoteGranted {
					fmt.Printf("%s 获得 %s 的投票\n", rf.address, rf.members[i])
					atomic.AddInt32(&votes, 1)
				} else {
					fmt.Printf("%s 未获得 %s 的投票\n", rf.address, rf.members[i])
				}
				if atomic.LoadInt32(&votes) > int32(n/2) {
					fmt.Printf("%s 获得过半的投票，成为leader\n", rf.address)
					rf.beLeader()
					send(rf.voteCh)
				}
			} else {
				fmt.Println("RequestVote返回结果失败")
			}
		}(i)
	}
}

func (rf *Raft) sendRequestVote(address string, args *RPC.RequestVoteArgs) (bool, *RPC.RequestVoteReply) {
	// RequestVote RPC 中的Client端
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		fmt.Println("拨号失败")
		fmt.Println(err1)
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			fmt.Println("关闭拨号失败")
			fmt.Println(err2)
		}
	}()
	client := RPC.NewRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err3 := client.RequestVote(ctx, args)
	if err3 != nil {
		fmt.Println("接受RequestVote结果失败:",err3)
		return false, reply
	}

	return true, reply
}

func (rf *Raft) RequestVote(ctx context.Context, args *RPC.RequestVoteArgs) (*RPC.RequestVoteReply, error) {
	// 方法实现端
	fmt.Printf("·····1····%s 收到投票请求：·········\n", rf.address)
	fmt.Printf("请求者信息：term：%d  index：%d\n", args.Term, args.LastLogIndex)
	fmt.Printf("接受者信息：term：%d  index：%d\n", rf.currentTerm, rf.getLastLogIndex())
	reply := &RPC.RequestVoteReply{VoteGranted:false}
	reply.Term = rf.currentTerm //用于candidate更新自己的current
	// 发送者：args-term
	// 接收者：rf-currentTerm
	// ????根据下面这行代码，已投票的candidate/follower发现自己的term过时后会成为follower（清空votedFor），那之前的投票会被收回吗？
	if rf.currentTerm < args.Term {
		// candidate1 发送RPC到 candidate2，candidate2发现自己的term过时了，candidate2立即变成follower，再判断要不要给candidate1投票
		//？ candidate1 发送RPC到 follower，follower发现的term过时，清空自己的votedFor，再判断要不要给candidate1投票
		fmt.Printf("%d term 过期，成为follower\n", rf.me)
		rf.beFollower(args.Term) // 待细究
	}
	if args.Term >= rf.currentTerm && (rf.votedFor == NULL || rf.votedFor == args.CandidateId) &&
		(args.LastLogTerm > rf.getLastLogTerm() ||
			(args.LastLogTerm == rf.getLastLogTerm() && args.LastLogIndex >= rf.getLastLogIndex())) {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		send(rf.voteCh)
	}
	if reply.VoteGranted {
		fmt.Printf("······2·····%s投出自己的票·······\n", rf.address)
	} else {
		fmt.Printf("······2·····%s拒绝投票·······\n", rf.address)
	}
	return reply, nil
}

/*
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
			for {
				appendEntries := rf.log[rf.nextIndex[idx]:] // 刚开始是空的，为了维持自己的leader身份
				entries, _ := json.Marshal(appendEntries) // 转码
				args := &RPC.AppendEntriesArgs{
					Term:          rf.currentTerm,
					LeaderId:      rf.me,
					PrevLogIndex:   rf.getPrevLogIndex(idx),
					PrevLogTerm:    rf.getPrevLogTerm(idx),
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
					rf.matchIndex[idx] = args.PrevLogIndex + int32(len(appendEntries)) // 刚开始日志还未复制，matchIndex=prevLogIndex
					rf.nextIndex[idx] = rf.matchIndex[idx] + 1
					rf.updateCommitIndex()
					return
				} else {
					rf.nextIndex[idx]--
				}
			}
		}(i)
	}

}*/


func (rf *Raft) startAppendEntries() {
	fmt.Println("startAppendLog")

	for i := 0; i < len(rf.members); i++ {
		if rf.address == rf.members[i] {
			continue
		}
		go func(idx int) {
			fmt.Printf("%s 向 %s append log, term :%d\n",rf.address, rf.members[idx], rf.currentTerm)
		}(i)
	}
}


func (rf *Raft) sendHeartBeat() {
	fmt.Printf("############ 发送心跳包 me:%d term:%d ############\n", rf.me, rf.currentTerm)
	n := len(rf.members)
	for i := 0; i < n; i++ {
		go func(idx int) {
			for {
				args := &RPC.AppendEntriesArgs{
					Term:          rf.currentTerm,
					LeaderId:      rf.me,
					PrevLogIndex:  rf.getPrevLogIndex(idx),
					PrevLogTerm:   rf.getPrevLogTerm(idx),
					Entries:       nil,
					LeaderCommit:  rf.commitIndex,
				}
				reply := rf.sendAppendEntries(rf.members[idx], args)
				if rf.role != Leader || rf.currentTerm != args.Term {
					return
				}
				if reply.Term > rf.currentTerm {
					rf.beFollower(reply.Term)
					return
				}
			}
		}(i)
	}
}

func (rf *Raft) updateCommitIndex() {
	n := len(rf.matchIndex)
	copyMatchIndex := make([]int32, n)
	copy(copyMatchIndex, rf.matchIndex)
	sort.Sort(IntSlice(copyMatchIndex))
	N := copyMatchIndex[n / 2] // 过半
	if N > rf.commitIndex && rf.log[N].term == rf.currentTerm {
		rf.commitIndex = N
		rf.updateLastApplied()
	}
}

func (rf *Raft) updateLastApplied() { // apply
	for rf.lastApplied < rf.commitIndex {
		rf.lastApplied++
		curEntry := rf.log[rf.lastApplied]
		cm := curEntry.Command
		if cm.Option == "Put" {
			rf.Persist.Put(cm.Key, cm.Value)
		}
		applyMsg := ApplyMsg{
			true,
			curEntry.Command,
			int(rf.lastApplied),
		}
		rf.applyCh <- applyMsg
	}
}

func (rf *Raft) sendAppendEntries(address string, args *RPC.AppendEntriesArgs) *RPC.AppendEntriesReply {
	// AppendEntries RPC 的Client端
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		fmt.Println(err1)
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			fmt.Println(err2)
		}
	}()
	client := RPC.NewRaftClient(conn)
	reply, err3 := client.AppendEntries(context.Background(), args)
	if err3 != nil {
		fmt.Println(err3)
	}
	return reply
}

func (rf *Raft) AppendEntries(ctx context.Context, args *RPC.AppendEntriesArgs) (*RPC.AppendEntriesReply, error) {
	// 发送者：args-Term
	// 接收者：rf-currentTerm
	reply := &RPC.AppendEntriesReply{}
	reply.Term = rf.currentTerm
	reply.Success = false
	if rf.currentTerm < args.Term { // 接收者发现自己的term过期了，更新term，转成follower
		rf.beFollower(args.Term)
	}
	rfLogLen := int32(len(rf.log))
	if args.Term < rf.currentTerm || rfLogLen <= args.PrevLogIndex ||
		rfLogLen > args.PrevLogIndex && rf.log[args.PrevLogIndex].term != args.PrevLogTerm { // 返回后发送者会转变成 follower
		return reply, nil
	}
	var newEntries []Entry
	err := json.Unmarshal(args.Entries, &newEntries)
	if err != nil {
		fmt.Println(err)
	}
	rf.log = rf.log[:args.PrevLogIndex+1]
	rf.log = append(rf.log, newEntries[0:]...)
	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = Min(args.LeaderCommit, rf.getLastLogIndex())
		rf.updateLastApplied()
	}
	reply.Success = true
	return reply, nil
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
	fmt.Println(rf.address, " become Candidate, new Term: ", rf.currentTerm)
	go rf.startElection()
}

func (rf *Raft) beFollower(term int32) {
	rf.role = Follower
	rf.votedFor = NULL
	rf.currentTerm = term
}

func (rf *Raft) beLeader() {
	rf.role = Leader
	n := len(rf.members)
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
	return rf.nextIndex[i]-1 //
}

func (rf *Raft) getPrevLogTerm(i int) int32 {
	prevLogIndex := rf.getPrevLogIndex(i)
	if prevLogIndex == 0 { //empty
		return -1
	}
	return rf.log[prevLogIndex].term
}

func (rf *Raft) GetState() (int32, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term := rf.currentTerm
	isLeader := rf.role == Leader
	return term, isLeader
}

func (s IntSlice) Len() int {
	return len(s)
}

func (s IntSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s IntSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func Min(a, b int32) int32 {
	if a > b {
		return b
	} else {
		return a
	}
}