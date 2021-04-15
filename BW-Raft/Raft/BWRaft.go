package main

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
	Term int32
	Command Op
}

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

type BWRaft struct {
	mu *sync.Mutex
	me int32
	address string
	members []string //其他成员，包括自己
	role State

	currentTerm int32
	votedFor int32
	electionTime time.Duration
	log []Entry

	commitIndex int32
	lastApplied int32

	nextIndex []int32
	matchIndex []int32

	voteCh chan bool
	appendLogCh chan bool
	killCh chan bool
	applyCh chan ApplyMsg
	replicateCh chan bool

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
func (rf *BWRaft) Start(command interface{}) (int32, int32, bool) {
	var index int32 = -1
	term := rf.currentTerm
	isLeader := rf.role == Leader
	if isLeader {
		index = rf.getLastLogIndex() + 1
		newEntry := Entry{
			Term:    rf.currentTerm,
			Command: command.(Op), //？
		}
		rf.log = append(rf.log, newEntry)
		fmt.Printf("Start()----新日志的Index：%d，term：%d，内容：%s, startAppendEntries\n", index, newEntry.Term, newEntry.Command)
		rf.startAppendEntries()

	}
	return index, term, isLeader
}

func (rf *BWRaft) registerServer(address string) {
	//BWRaft内部的Server端
	server := grpc.NewServer()
	RPC.RegisterBWRaftServer(server, rf)
	lis, err1 := net.Listen("tcp", address)
	if err1 != nil {
		fmt.Println(err1)
	}
	err2 := server.Serve(lis)
	if err2 != nil {
		fmt.Println(err2)
	}
}


func (rf *BWRaft) IsConnect (ctx context.Context, args *RPC.ConnectArgs) (*RPC.ConnectReply, error) {
	// Connect RPC的server端
	reply := &RPC.ConnectReply{Success:true}
	return reply, nil
}

func MakeBWBWRaft(address string, members []string, persist *PERSISTER.Persister, mu *sync.Mutex, ) *BWRaft {
	BWRaft := &BWRaft{}
	BWRaft.address = address
	BWRaft.me = getMe(address)
	BWRaft.members = members
	BWRaft.Persist = persist
	BWRaft.mu = mu
	n := len(BWRaft.members)
	fmt.Printf("当前节点:%s, rf.me=%d, 所有成员地址：\n", BWRaft.address, BWRaft.me)
	for i := 0; i < n; i++ {
		fmt.Println(BWRaft.members[i])
	}
	BWRaft.init()
	return BWRaft
}


func (rf *BWRaft) init() {
	rf.role = Follower
	rf.currentTerm = 0
	rf.votedFor = NULL
	rf.log = make([]Entry, 1) // 日志索引从1开始
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.voteCh = make(chan bool, 1)
	rf.appendLogCh = make(chan bool, 1)
	rf.killCh = make(chan bool, 1)
	rf.applyCh = make(chan ApplyMsg, 1)

	heartbeatTime := time.Duration(150) * time.Millisecond

	go func() {
		for {
			select {
			case <-rf.killCh:
				return
			default:
			}
			//electionTime := time.Duration(rand.Intn(350)+500) * time.Millisecond
			state := rf.role
			switch state {
			case Follower, Candidate:
				select {
				case <-rf.voteCh:
				case <-rf.appendLogCh:
				//以上两个Ch表示Follower/Candidate已经在对相应的请求进行了响应，否则从程序运行开始就进行选举超时计时了
				case <-time.After(rf.electionTime):
					fmt.Println("######## time.After(electionTime) #######")
					rf.beCandidate()
				}
			case Leader:
				rf.startAppendEntries()
				fmt.Printf("--------sleep heartbeat time------\n")
				time.Sleep(heartbeatTime)
			}
		}
	}()

	go rf.registerServer(rf.address)

}


func (rf *BWRaft) startElection() {
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
		if rf.role != Candidate {
			fmt.Println("Candidate 角色变更")
			return
		}
		if rf.members[i] == rf.address {
			continue
		}
		go func(idx int) {
			if rf.role != Candidate {
				return
			}
			fmt.Printf("%s --> %s RequestVote RPC\n", rf.address, rf.members[idx])
			ret, reply := rf.sendRequestVote(rf.members[idx], args) //一定要有ret
			if ret {
				if reply.Term > rf.currentTerm { // 此Candidate的term过时
					fmt.Println(rf.address, " 的term过期，转成follower")
					rf.beFollower(reply.Term)
					return
				}
				if rf.role != Candidate || rf.currentTerm != args.Term{ // 有其他candidate当选了leader或者不是在最新Term（rf.currentTerm）进行的投票
					return
				}
				if reply.VoteGranted {
					fmt.Printf("%s 获得 %s 的投票\n", rf.address, rf.members[idx])
					atomic.AddInt32(&votes, 1)
				} else {
					fmt.Printf("%s 未获得 %s 的投票\n", rf.address, rf.members[idx])
				}
				if atomic.LoadInt32(&votes) > int32(n/2) {
					rf.beLeader()
					send(rf.voteCh)
				}
			} else {
				fmt.Println("RequestVote返回结果失败")
			}
		}(i)
	}
}




func (rf *BWRaft) sendRequestVote(address string, args *RPC.RequestVoteArgs) (bool, *RPC.RequestVoteReply) {
	// RequestVote RPC 中的Client端
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
	client := RPC.NewBWRaftClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err3 := client.RequestVote(ctx, args)
	if err3 != nil {
		fmt.Println("接受RequestVote结果失败:",err3)
		return false, reply
	}
	return true, reply
}



func (rf *BWRaft) RequestVote(ctx context.Context, args *RPC.RequestVoteArgs) (*RPC.RequestVoteReply, error) {
	// 方法实现端
	// 发送者：args-term
	// 接收者：rf-currentTerm
	fmt.Printf("\n·····1····%s 收到投票请求：·········\n", rf.address)
	reply := &RPC.RequestVoteReply{VoteGranted:false}
	reply.Term = rf.currentTerm //用于candidate更新自己的current
	if rf.currentTerm < args.Term { //旧term时的投票已经无效，现在只关心最新term时的投票情况
		fmt.Printf("接收者的term: %d < 发送者的term: %d，成为follower\n", rf.currentTerm, args.Term)
		rf.beFollower(args.Term) // 清空votedFor
	}
	fmt.Printf("请求者：term：%d  index：%d lastTerm:%d\n", args.Term, args.LastLogIndex, args.LastLogTerm)
	fmt.Printf("接收者：term：%d  index：%d lastTerm:%d vote:%d\n", rf.currentTerm, rf.getLastLogIndex(), rf.getLastLogTerm(), rf.votedFor)
	if args.Term >= rf.currentTerm && (rf.votedFor == NULL || rf.votedFor == args.CandidateId) &&
		(args.LastLogTerm > rf.getLastLogTerm() ||
			(args.LastLogTerm == rf.getLastLogTerm() && args.LastLogIndex >= rf.getLastLogIndex())) {
		reply.VoteGranted = true
		rf.votedFor = args.CandidateId
		rf.role = Follower
		send(rf.voteCh)
	}
	if reply.VoteGranted {
		fmt.Printf("·····2····请求者term:%d···%s投出自己的票·······\n\n", args.Term, rf.address)
	} else {
		fmt.Printf("·····2····请求者term:%d···%s拒绝投票·······\n\n", args.Term, rf.address)
	}

	return reply, nil
}



func (rf *BWRaft) startAppendEntries() {
	fmt.Printf("############ 开始日志追加 me:%d term:%d ############\n", rf.me, rf.currentTerm)
	n := len(rf.members)
	for i := 0; i < n; i++ {
		if rf.role != Leader {
			return
		}
		if rf.members[i] == rf.address {
			continue
		}
		go func(idx int) {
			for { //因为会遇到冲突使nextIndex--，所以不能只执行一次
				appendEntries := rf.log[rf.nextIndex[idx]:] // 刚开始是空的，为了维持自己的leader身份
				if len(appendEntries) > 0 {
					fmt.Printf("待追加日志长度appendEntries：%d 信息：term：%d cm:%s\n", len(appendEntries), appendEntries[0].Term, appendEntries[0].Command)
				}
				entries, _ := json.Marshal(appendEntries) // 转码
				args := &RPC.AppendEntriesArgs{ //prevLogIndex和prevLogTerm会因为冲突而改变，所以放在了for循环内
					Term:          rf.currentTerm,
					LeaderId:      rf.me,
					PrevLogIndex:   rf.getPrevLogIndex(idx),
					PrevLogTerm:    rf.getPrevLogTerm(idx),
					Entries:       entries,
					LeaderCommit:  rf.commitIndex,
				}
				if len(appendEntries) == 0 {
					fmt.Printf("%s --> %s  HeartBeat RPC\n", rf.address, rf.members[idx])
				} else {
					fmt.Printf("%s --> %s  AppendEntries RPC\n", rf.address, rf.members[idx])
				}
				ret, reply := rf.sendAppendEntries(rf.members[idx], args)
				if ret {
					if reply.Term > rf.currentTerm {
						rf.beFollower(reply.Term)
						return
					}
					if rf.role != Leader || reply.Term != rf.currentTerm { //只处理最新term（rf.currentTerm）的数据
						return
					}
					if reply.Success {
						rf.matchIndex[idx] = args.PrevLogIndex + int32(len(appendEntries)) // 刚开始日志还未复制，matchIndex=prevLogIndex
						rf.nextIndex[idx] = rf.matchIndex[idx] + 1
						if len(appendEntries) == 0 {
							fmt.Printf("向 %s 发送心跳包成功，nextIndex[%s]=%d\n", rf.members[idx], rf.members[idx], rf.nextIndex[idx])
						} else {
							fmt.Printf("向 %s 日志追加成功，nextIndex[%s]=%d\n", rf.members[idx], rf.members[idx], rf.nextIndex[idx])
						}
						rf.updateCommitIndex()
						return
					} else {
						rf.nextIndex[idx]--
						fmt.Printf("日志不匹配，更新后nextIndex[%s]=%d\n", rf.members[idx], rf.nextIndex[idx])
					}
				}
			}
		}(i)
	}

}



/*
func (rf *BWRaft) startAppendEntries() {
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
*/


func (rf *BWRaft) updateCommitIndex()  { // 只由leader调用
	n := len(rf.matchIndex)
	for i := 0; i  < n; i++ {
		fmt.Printf("matchIndex[%s]=%d\n", rf.members[i], rf.matchIndex[i])
	}
	copyMatchIndex := make([]int32, n)
	copy(copyMatchIndex, rf.matchIndex)
	sort.Sort(IntSlice(copyMatchIndex))
	N := copyMatchIndex[n / 2] // 过半
	fmt.Printf("过半的commitIndex：%d\n", N)
	if N > rf.commitIndex && rf.log[N].Term == rf.currentTerm {
		rf.commitIndex = N
		fmt.Printf("new LeaderCommit：%d\n", rf.commitIndex)
		rf.updateLastApplied()
	}
}

func (rf *BWRaft) updateLastApplied()  { // apply
	fmt.Printf("updateLastApplied()---lastApplied: %d, commitIndex: %d\n", rf.lastApplied, rf.commitIndex)
	for rf.lastApplied < rf.commitIndex { // 0 0
		rf.lastApplied++
		curEntry := rf.log[rf.lastApplied]
		cm := curEntry.Command
		if cm.Option == "write" {
			rf.Persist.Put(cm.Key, cm.Value)
			fmt.Printf("Write 命令：%s-%s 被apply\n", cm.Key, cm.Value)
		} else if cm.Option == "read" {
			fmt.Printf("Read 命令：key:%s 被apply\n", cm.Key)
		}
		applyMsg := ApplyMsg{
			true,
			curEntry.Command,
			int(rf.lastApplied),
		}
		//rf.applyCh <- applyMsg
		applyRequest(rf.applyCh, applyMsg)
	}
}

func applyRequest(ch chan ApplyMsg, msg ApplyMsg) {
	select {
	case <- ch:
	default:
	}
	ch <- msg
}

func (rf *BWRaft) sendAppendEntries(address string, args *RPC.AppendEntriesArgs) (bool, *RPC.AppendEntriesReply) {
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
	client := RPC.NewBWRaftClient(conn)
	reply, err3 := client.AppendEntries(context.Background(), args)
	if err3 != nil {
		fmt.Println(err3)
		return false, reply
	}
	return true, reply
}

func (rf *BWRaft) AppendEntries(ctx context.Context, args *RPC.AppendEntriesArgs) (*RPC.AppendEntriesReply, error) {
	// 发送者：args-Term
	// 接收者：rf-currentTerm
	fmt.Printf("\n~~~~~~1~~~~~进入AppendEntries~~~~~~~\n")
	reply := &RPC.AppendEntriesReply{}
	reply.Term = rf.currentTerm
	reply.Success = false
	if rf.currentTerm < args.Term { // 接收者发现自己的term过期了，更新term，转成follower
		fmt.Printf("接收者的term: %d < 发送者的term: %d，成为follower\n", rf.currentTerm, args.Term)
		rf.beFollower(args.Term)
	}
	rfLogLen := int32(len(rf.log))
	fmt.Printf("请求者：term：%d  pIndex：%d prevTerm:%d\n", args.Term, args.PrevLogIndex, args.PrevLogTerm)
	if rfLogLen > args.PrevLogIndex {
		fmt.Printf("接收者：term：%d  loglen：%d prevTerm:%d\n", rf.currentTerm, len(rf.log), rf.log[args.PrevLogIndex].Term)
	} else {
		fmt.Printf("接收者：term：%d  loglen：%d prevTerm:NULL\n", rf.currentTerm, len(rf.log))
	}
	if args.Term < rf.currentTerm || rfLogLen <= args.PrevLogIndex ||
		rfLogLen > args.PrevLogIndex && rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
		fmt.Printf("~~~~~~2~~~~~日志冲突~~~~~~~\n\n")
		return reply, nil
	}
	fmt.Printf("~~~~~~2~~~~~日志匹配~~~~~~~\n\n")
	var newEntries []Entry
	err := json.Unmarshal(args.Entries, &newEntries)
	if err != nil {
		fmt.Println(err)
	}
	if len(newEntries) > 0 {
		fmt.Printf("日志匹配，待追加日志长度：%d 第一个内容：term：%d command：%s\n", len(newEntries), newEntries[0].Term, newEntries[0].Command)
	} else {
		fmt.Printf("心跳包，待追加日志长度为0\n")
	}
	rf.log = rf.log[:args.PrevLogIndex+1]
	rf.log = append(rf.log, newEntries[0:]...)
	fmt.Printf("############## 更新后的日志 ###########\n")
	for i := 0; i < len(rf.log); i++ {
		fmt.Printf("index:%d term:%d command:%s\n", i, rf.log[i].Term, rf.log[i].Command)
	}
	if args.LeaderCommit > rf.commitIndex { // not 0 > 0
		rf.commitIndex = Min(args.LeaderCommit, rf.getLastLogIndex())
		rf.updateLastApplied() //在各自的存储引擎中存储相关的内容，在本此实现中，集群成员共用一个存储引擎，此处可以生省略。
	}
	send(rf.appendLogCh)
	reply.Success = true
	return reply, nil
}

func send(ch chan bool) {
	select {
	case <-ch:
	default:
	}
	ch <- true
}


func (rf *BWRaft) beCandidate() {
	rf.role = Candidate
	rf.currentTerm++
	rf.votedFor = rf.me
	rf.electionTime = time.Duration(rand.Intn(350)+500) * time.Millisecond
	fmt.Println(rf.address, " become Candidate, new Term: ", rf.currentTerm)
	go rf.startElection()
}


func (rf *BWRaft) beFollower(term int32) {
	rf.role = Follower
	rf.votedFor = NULL
	rf.currentTerm = term
}


func (rf *BWRaft) beLeader() {
	if rf.role != Candidate {
		return
	}
	rf.role = Leader
	n := len(rf.members)
	rf.nextIndex = make([]int32, n)
	rf.matchIndex = make([]int32, n)
	for i := 0; i < n; i++ {
		rf.nextIndex[i] = rf.getLastLogIndex()+1
		fmt.Printf("beLeader---NextIndex[%s] = %d\n", rf.members[i], rf.nextIndex[i])
	}
	fmt.Println(rf.address, "---------------become LEADER--------------", rf.currentTerm)

}


func (rf *BWRaft) getLastLogIndex() int32 {
	// 数组下标 0 1 2 3 4
	// len=5, 最新日志索引Index=4，在数组中的下标也为4
	return int32(len(rf.log)-1) // empty, =0(最新日志索引)
}

func (rf *BWRaft) getLastLogTerm() int32 {
	index := rf.getLastLogIndex()
	return rf.log[index].Term
}

func (rf *BWRaft) getPrevLogIndex(i int) int32 {
	// 每个peer对应的nextIndex会变，其相应的prevLogIndex也会变
	fmt.Printf("NextIndex[%s] = %d, prevLogIndex=%d\n", rf.members[i], rf.nextIndex[i], rf.nextIndex[i]-1)
	return rf.nextIndex[i]-1 //
}

func (rf *BWRaft) getPrevLogTerm(i int) int32 {
	prevLogIndex := rf.getPrevLogIndex(i)
	fmt.Printf("########prevLogIndex:%d########\n", prevLogIndex)
	fmt.Printf("########prevLogIndexTerm:%d########\n", rf.log[prevLogIndex].Term)
	return rf.log[prevLogIndex].Term
}

func (rf *BWRaft) GetState() (int32, bool) {
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