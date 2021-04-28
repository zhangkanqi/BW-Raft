package Raft

import (
	"../RPC"
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
	Secretary
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
	members []string //其他成员，包括自己(不是sec和obs)
	secretaries []string
	role State

	detectAdd []string
	byzantine []string
	reception []string
	suspicion []string
	cntSuspicion map[string]int

	currentTerm int32
	votedFor int32
	electionTime time.Duration
	log []Entry

	commitIndex int32
	lastApplied int32

	NextIndexf []int32
	MatchIndexf []int32
	nextIndexs []int32
	matchIndexs []int32

	voteCh chan bool
	appendLogCh chan bool
	killCh chan bool
	applyCh chan ApplyMsg
	replicateCh chan bool
	secretaryAppenEntriesFromLeaderCh chan bool
	detectCh chan bool
	updateByzCh chan bool

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
		//rf.startAppendEntries()
		isLeader = rf.leaderAppendEntriesToSecretaries()
		if !isLeader {
			return index, term, isLeader
		}
		//这里不需要写rf.startAppendEntries()，因为init()里面Leader会每隔heartbeat时间执行rf.startAppendEntries()
	}
	return index, term, isLeader
}

// 5000 BWRaft
func (rf *BWRaft) registerServer(address string) {
	fmt.Println("开始注册内部服务器", address)
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

// 50002 Connect
func (rf *BWRaft) registerServer2(address string) {
	//BWRaft内部的Server端
	server := grpc.NewServer()
	RPC.RegisterGetValueServer(server, rf)
	lis, err1 := net.Listen("tcp", address)
	if err1 != nil {
		fmt.Println(err1)
	}
	err2 := server.Serve(lis)
	if err2 != nil {
		fmt.Println(err2)
	}
}

func (rf *BWRaft) GetValue (ctx context.Context, args *RPC.GetValueArgs) (*RPC.GetValueReply, error) {
	// GetValue RPC的server端
	reply := &RPC.GetValueReply{
		Success:	true,
		Value:		rf.Persist.Get(args.Key),
	}
	return reply, nil
}

func MakeSecretary(address string, members []string, persist *PERSISTER.Persister, mu *sync.Mutex) {
	BWRaft := &BWRaft{}
	BWRaft.address = address
	BWRaft.me = getMe(address)
	BWRaft.members = members
	BWRaft.Persist = persist
	BWRaft.mu = mu
	fmt.Printf("当前节点Secretary:%s, 所有成员地址：\n", BWRaft.address)
	for i, j := range BWRaft.members {
		fmt.Println(i, j)
	}
	BWRaft.initSecretary()
}

func MakeBWRaft(address string, members []string, secretaries []string, detectAdd []string, persist *PERSISTER.Persister, mu *sync.Mutex) *BWRaft {
	BWRaft := &BWRaft{}
	BWRaft.address = address
	BWRaft.me = getMe(address)
	BWRaft.members = members
	BWRaft.secretaries = secretaries
	BWRaft.detectAdd = detectAdd
	BWRaft.Persist = persist
	BWRaft.mu = mu
	fmt.Printf("当前节点:%s, rf.me=%d, 所有成员地址：\n", BWRaft.address, BWRaft.me)
	for i, j := range BWRaft.members {
		fmt.Println(i, j)
	}
	fmt.Printf("所有secretary：\n")
	for i, j := range BWRaft.secretaries {
		fmt.Println(i, j)
	}
	BWRaft.init()
	return BWRaft
}

func (rf *BWRaft) initSecretary() {
	rf.role = Secretary
	rf.currentTerm = 0
	rf.log = make([]Entry, 1) // 日志索引从1开始
	rf.commitIndex = 0
	rf.lastApplied = 0

	rf.appendLogCh = make(chan bool, 1)
	rf.applyCh = make(chan ApplyMsg, 1)
	rf.secretaryAppenEntriesFromLeaderCh = make(chan bool, 1)

	go func() {
		for {
			fmt.Println("身份：", rf.role)
			switch rf.role {
			case Secretary:
				select {
				case <- rf.secretaryAppenEntriesFromLeaderCh:
					rf.secretaryAppendEntriesToFollower()
				}
			}
		}
	}()

	//5000 Raft 内部服务
	go rf.registerServer(rf.address)
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
	go rf.registerServer2(rf.address+"2")
	go rf.registerServer3(rf.address+"3")
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

// leader发起
func (rf *BWRaft) leaderAppendEntriesToSecretaries() bool { //返回isLeader
	n := len(rf.secretaries)
	for i := 0; i < n; i++ {
		if rf.role != Leader {
			return false
		}
		go func(idx int) {
			for {
				appendEntries := rf.log[rf.nextIndexs[idx]:] // 记得去更新NextIndexf，包含follower和secretary
				if len(appendEntries) > 0 {
					fmt.Printf("待发送日志长度appendEntries：%d 信息：term：%d cm:%s\n", len(appendEntries), appendEntries[0].Term, appendEntries[0].Command)
				}
				entries, _ := json.Marshal(appendEntries) // 转码
				Nextf, _ := json.Marshal(rf.NextIndexf)
				Matchf, _ := json.Marshal(rf.MatchIndexf)
				args := &RPC.AppendEntriesArgs{ //prevLogIndex和prevLogTerm会因为冲突而改变，所以放在了for循环内
					Term:          rf.currentTerm,
					LeaderId:      rf.me,
					PrevLogIndex:   rf.getPrevLogIndexs(idx),
					PrevLogTerm:    rf.getPrevLogTerms(idx),
					Entries:       entries,
					LeaderCommit:  rf.commitIndex,
					LastApplied:	rf.lastApplied,
					NextIndexf:		Nextf,
					MatchIndexf:	Matchf,
					Role: 			"Leader",
					Address:		rf.address,
				}
				fmt.Printf("Leader %s --> Secretary %s  AppendEntries RPC\n", rf.address, rf.secretaries[idx])
				ret, reply := rf.sendAppendEntries(rf.secretaries[idx], args)
				if ret {
					fmt.Printf("(reply) Secretary.Term:%d Leader.term:%d\n", reply.Term, rf.currentTerm)
					if reply.Term > rf.currentTerm {
						rf.beFollower(reply.Term)
						return
					}
					if rf.role != Leader || reply.Term != rf.currentTerm { //只处理最新term（rf.currentTerm）的数据
						return
					}
					if reply.Success {
						rf.matchIndexs[idx] = args.PrevLogIndex + int32(len(appendEntries)) // 刚开始日志还未复制，MatchIndexf=prevLogIndex
						rf.nextIndexs[idx] = rf.matchIndexs[idx] + 1
						fmt.Printf("Leader-%s 向 Secretary-%s 追加日志成功，nextIndexs[%s]=%d\n", rf.address, rf.secretaries[idx], rf.secretaries[idx], rf.nextIndexs[idx])
						return
					} else {
						rf.nextIndexs[idx]--
						fmt.Printf("日志不匹配，更新后nextIndexs[%s]=%d\n", rf.secretaries[idx], rf.nextIndexs[idx])
					}
				} else {
					fmt.Printf("接收Secretary-%s的append返回信息失败\n", rf.secretaries[idx])
				}
			}
		}(i)
	}
	return true
}

//secretary
func (rf *BWRaft) secretaryAppendEntriesToFollower() {
	fmt.Printf("############ Secretary %s 开始向Follower追加日志 ############\n", rf.address)
	n := len(rf.members)
	for i, j := range rf.NextIndexf {
		fmt.Printf("Secretary %s 的nextIndexf[%s]=%d\n", rf.address, rf.members[i], j)
	}
	for i, j := range rf.MatchIndexf {
		fmt.Printf("Secretary %s 的matchIndexf[%s]=%d\n", rf.address, rf.members[i], j)
	}
	for i := 0; i < n; i++ {
		go func(idx int) {
			for { //因为会遇到冲突使NextIndexf--，所以不能只执行一次
				appendEntries := rf.log[rf.NextIndexf[idx]:] // 刚开始是空的，为了维持自己的leader身份
				fmt.Printf("待追加日志长度appendEntries：%d 信息：term：%d cm:%s\n", len(appendEntries), appendEntries[0].Term, appendEntries[0].Command)
				entries, _ := json.Marshal(appendEntries) // 转码
				args := &RPC.AppendEntriesArgs{ //prevLogIndex和prevLogTerm会因为冲突而改变，所以放在了for循环内
					Term:          rf.currentTerm,
					LeaderId:      rf.me,
					PrevLogIndex:   rf.getPrevLogIndexf(idx),
					PrevLogTerm:    rf.getPrevLogTermf(idx),
					Entries:       entries,
					LeaderCommit:  rf.commitIndex,
					Role: 			"Secretary",
					Address: 	rf.address,
				}
				fmt.Printf("Secretary %s --> Follower %s  AppendEntries RPC\n", rf.address, rf.members[idx])
				ret, reply := rf.sendAppendEntries(rf.members[idx], args)
				if ret {
					if reply.Term > rf.currentTerm {
						return
					}
					if rf.role != Secretary || reply.Term != rf.currentTerm { //只处理最新term（rf.currentTerm）的数据
						return
					}
					if reply.Success {
						rf.MatchIndexf[idx] = args.PrevLogIndex + int32(len(appendEntries)) // 刚开始日志还未复制，MatchIndexf=prevLogIndex
						rf.NextIndexf[idx] = rf.MatchIndexf[idx] + 1
						if len(appendEntries) == 0 {
							fmt.Printf("Secretary-%s 向 Follower-%s发送心跳包成功，NextIndexf[%s]=%d\n", rf.address, rf.members[idx], rf.members[idx], rf.NextIndexf[idx])
						} else {
							fmt.Printf("Secretary-%s 向 Follower-%s日志追加成功，NextIndexf[%s]=%d\n", rf.address, rf.members[idx], rf.members[idx], rf.NextIndexf[idx])
						}
						rf.updateCommitIndex()
						return
					} else {
						rf.NextIndexf[idx]--
						fmt.Printf("日志不匹配，更新后NextIndexf[%s]=%d\n", rf.members[idx], rf.NextIndexf[idx])
					}
				}
			}
		}(i)
	}
}


func (rf *BWRaft) startAppendEntries() {
	//fmt.Printf("############ 开始日志追加 me:%d term:%d ############\n", rf.me, rf.currentTerm)
	n := len(rf.members)
	for i := 0; i < n; i++ {
		if rf.role != Leader {
			return
		}
		if rf.members[i] == rf.address {
			continue
		}
		go func(idx int) {
			for { //因为会遇到冲突使NextIndexf--，所以不能只执行一次
				appendEntries := rf.log[rf.NextIndexf[idx]:] // 刚开始是空的，为了维持自己的leader身份
				if len(appendEntries) > 0 {
					fmt.Printf("待追加日志长度appendEntries：%d 信息：term：%d cm:%s\n", len(appendEntries), appendEntries[0].Term, appendEntries[0].Command)
				}
				entries, _ := json.Marshal(appendEntries) // 转码
				args := &RPC.AppendEntriesArgs{ //prevLogIndex和prevLogTerm会因为冲突而改变，所以放在了for循环内
					Term:          rf.currentTerm,
					LeaderId:      rf.me,
					PrevLogIndex:   rf.getPrevLogIndexf(idx),
					PrevLogTerm:    rf.getPrevLogTermf(idx),
					Entries:       entries,
					LeaderCommit:  rf.commitIndex,
					Role:		"Leader",
					Address:		rf.address,
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
						rf.MatchIndexf[idx] = args.PrevLogIndex + int32(len(appendEntries)) // 刚开始日志还未复制，MatchIndexf=prevLogIndex
						rf.NextIndexf[idx] = rf.MatchIndexf[idx] + 1
						if len(appendEntries) == 0 {
							fmt.Printf("Leader-%s向Follower-%s 发送心跳包成功，NextIndexf[%s]=%d\n", rf.address, rf.members[idx], rf.members[idx], rf.NextIndexf[idx])
						} else {
							fmt.Printf("Leader-%s向Follower-%s 日志追加成功，NextIndexf[%s]=%d\n", rf.address, rf.members[idx], rf.members[idx], rf.NextIndexf[idx])
						}
						rf.updateCommitIndex()
						return
					} else {
						rf.NextIndexf[idx]--
						fmt.Printf("日志不匹配，更新后NextIndexf[%s]=%d\n", rf.members[idx], rf.NextIndexf[idx])
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


func (rf *BWRaft) updateCommitIndex()  {
	n := len(rf.MatchIndexf)
	for i := 0; i  < n; i++ {
		fmt.Printf("MatchIndexf[%s]=%d\n", rf.members[i], rf.MatchIndexf[i])
	}
	copyMatchIndexf := make([]int32, n)
	copy(copyMatchIndexf, rf.MatchIndexf)
	sort.Sort(IntSlice(copyMatchIndexf))
	N := copyMatchIndexf[n / 2] // 过半
	fmt.Printf("过半的commitIndex：%d\n", N)
	if N > rf.commitIndex && rf.log[N].Term == rf.currentTerm {
		rf.commitIndex = N
		fmt.Printf("new LeaderCommit：%d\n", rf.commitIndex)
		rf.updateLastApplied()
	}
}

//Leader/secretary
func (rf *BWRaft) updateLastApplied()  { // apply
	fmt.Printf("updateLastApplied()---lastApplied: %d, commitIndex: %d\n", rf.lastApplied, rf.commitIndex)
	for rf.lastApplied < rf.commitIndex { // 0 0
		rf.lastApplied++
		curEntry := rf.log[rf.lastApplied]
		cm := curEntry.Command
		if cm.Option == "write" {
			rf.Persist.Put(cm.Key, cm.Value)
			fmt.Printf("\n%d-%s：Write 命令：%s-%s 被apply\n\n", rf.role, rf.address, cm.Key, cm.Value)
		} else if cm.Option == "read" {
			fmt.Printf("\n%d-%s：Read 命令：key:%s 被apply\n\n", rf.role, rf.address, cm.Key)
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

//Secretary、Follower
func (rf *BWRaft) AppendEntries(ctx context.Context, args *RPC.AppendEntriesArgs) (*RPC.AppendEntriesReply, error) {
	// 发送者：args-Term
	// 接收者：rf-currentTerm
	fmt.Printf("\n~~~~~~1~~~~~收到%s-%s的AppendEntries~~~~~~~\n", args.Role, args.Address)
	reply := &RPC.AppendEntriesReply{}
	reply.Term = rf.currentTerm
	reply.Success = false
	if rf.currentTerm < args.Term { // 接收者发现自己的term过期了，更新term，转成follower
		if rf.role != Secretary {
			fmt.Printf("接收者的term: %d < 发送者的term: %d，成为follower\n", rf.currentTerm, args.Term)
			rf.beFollower(args.Term)
		} else { // 是Secretary
			rf.currentTerm = args.Term //省略发心跳包的步骤
			reply.Term = rf.currentTerm
		}
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

	//用于secretary接收leader的RPC，更新信息，使secretary与leader同步
	if rf.role == Secretary {
		NextIndexf := make([]int32, len(rf.members))
		MatchIndexf := make([]int32, len(rf.members))
		err = json.Unmarshal(args.NextIndexf, &NextIndexf)
		if err != nil {
			fmt.Println(err)
		}
		err = json.Unmarshal(args.MatchIndexf, &MatchIndexf)
		if err != nil {
			fmt.Println(err)
		}
		rf.NextIndexf = NextIndexf
		rf.MatchIndexf = MatchIndexf
		for i, j := range rf.NextIndexf {
			fmt.Printf("转码后的rf.NextIndexf[%s]=%d:\n", rf.members[i], j)
		}
		rf.commitIndex = args.LeaderCommit
		rf.lastApplied = args.LastApplied
	}

	if args.LeaderCommit > rf.commitIndex { // not 0 > 0
		rf.commitIndex = Min(args.LeaderCommit, rf.getLastLogIndex())
		rf.updateLastApplied() //follower在自己存储空间内apply
	}
	send(rf.appendLogCh)
	if rf.role == Secretary {
		send(rf.secretaryAppenEntriesFromLeaderCh) // Leader成功向该secretary追加日志
		fmt.Printf("Secretary %s 接收Leader appendEntries成功\n\n", rf.address)
	}
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
	//rf.Detector()
	//rf.broadcastByzAndSus()
	rf.startElection()
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
	rf.NextIndexf = make([]int32, n)
	rf.MatchIndexf = make([]int32, n)
	for i := 0; i < n; i++ {
		rf.NextIndexf[i] = rf.getLastLogIndex()+1
		fmt.Printf("beLeader---NextIndexf[%s] = %d\n", rf.members[i], rf.NextIndexf[i])
	}

	m := len(rf.secretaries)
	rf.nextIndexs = make([]int32, m)
	rf.matchIndexs = make([]int32, m)
	for i := 0; i < m; i++ {
		rf.nextIndexs[i] = rf.getLastLogIndex()+1
		fmt.Printf("beLeader---nextIndexs[%s] = %d\n", rf.secretaries[i], rf.nextIndexs[i])
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

func (rf *BWRaft) getPrevLogIndexf(i int) int32 {
	// 每个peer对应的NextIndexf会变，其相应的prevLogIndex也会变
	fmt.Printf("NextIndexf[%s] = %d, prevLogIndexf=%d\n", rf.members[i], rf.NextIndexf[i], rf.NextIndexf[i]-1)
	return rf.NextIndexf[i]-1 //
}

func (rf *BWRaft) getPrevLogIndexs(i int) int32 {
	fmt.Printf("nextIndexs[%s] = %d, prevLogIndexs=%d\n", rf.secretaries[i], rf.nextIndexs[i], rf.nextIndexs[i]-1)
	return rf.nextIndexs[i]-1 //
}

func (rf *BWRaft) getPrevLogTermf(i int) int32 {
	prevLogIndex := rf.getPrevLogIndexf(i)
	fmt.Printf("########prevLogIndexf:%d########\n", prevLogIndex)
	fmt.Printf("########prevLogIndexTermf:%d########\n", rf.log[prevLogIndex].Term)
	return rf.log[prevLogIndex].Term
}

func (rf *BWRaft) getPrevLogTerms(i int) int32 {
	prevLogIndex := rf.getPrevLogIndexs(i)
	fmt.Printf("########prevLogIndexs:%d########\n", prevLogIndex)
	fmt.Printf("########prevLogIndexTerms:%d########\n", rf.log[prevLogIndex].Term)
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

// 50003 用于Detec Byzantin
func (rf *BWRaft) registerServer3(address string) {
	//BWRaft内部的Server端
	server := grpc.NewServer()
	RPC.RegisterDetectServer(server, rf)
	lis, err1 := net.Listen("tcp", address)
	if err1 != nil {
		fmt.Println(err1)
	}
	err2 := server.Serve(lis)
	if err2 != nil {
		fmt.Println(err2)
	}
}

func (rf *BWRaft) DetectByzantine(ctx context.Context, args *RPC.DetectArgs) (*RPC.DetectReply, error) {
	reply := &RPC.DetectReply{
		Address:rf.address,
		Value:args.Value,
		Success:true,
	}
	return reply, nil
}


func (rf *BWRaft) initDetector() {
	rf.byzantine = []string{}
	rf.cntSuspicion = map[string]int{}
	rf.suspicion = []string{}
	rf.reception = []string{}
	fmt.Printf("init之后：byz:%d cnt:%d sus:%d rec:%d\n", len(rf.byzantine), len(rf.cntSuspicion), len(rf.suspicion), len(rf.reception))
}

func (rf *BWRaft) sendDetection(address string, args *RPC.DetectArgs) (bool, *RPC.DetectReply) {
	// Detect RPC 中的Client端
	fmt.Printf("向%s拨号\n", address)
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
	client := RPC.NewDetectClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	fmt.Printf("-----\n")
	defer cancel()
	reply, err3 := client.DetectByzantine(ctx, args)
	if reply == nil {
		fmt.Println("接收DetectByzantine结果失败:",err3)
		return false, reply
	} else {
		fmt.Printf("接收到了Detect的消息\n")
	}
	return true, reply
}

func (rf *BWRaft) Detector() {
	rf.initDetector()
	var reception int32  = 0
	m := len(rf.detectAdd)
	quorum := int32(m*2/3)
	rec := make(map[string]bool)
	args := &RPC.DetectArgs{Value:rf.me}
	for _, address := range rf.detectAdd {
		if address ==  rf.address {
			continue
		}
		if atomic.LoadInt32(&reception) >= quorum {
			break
		}
		go func (add string) {
			fmt.Printf("%s向%s发送检测消息\n", rf.address, add)
			ret, reply := rf.sendDetection(add+"3", args)
			if ret {
				rf.reception = append(rf.reception, add)
				rec[add] = true
				fmt.Printf("%s接收到%s返回的检测消息：%d %s\n", rf.address, add, reply.Value, reply.Address)
				atomic.AddInt32(&reception, 1)
				//检测出拜占庭行为
				if (reply.Value != args.Value || reply.Address != add) && reply.Success == true {
					rf.byzantine = append(rf.byzantine, add)
				}
				return
			} else {
				fmt.Printf("%s未成功接收%s的DetectByzantine返回消息\n", rf.address, add)
				return
			}
		}(address)
	}
	go rf.addSuspicion(rec)
	return
}

func (rf *BWRaft) addSuspicion(rec map[string]bool) {
	for _, address := range rf.detectAdd {
		if _, ok := rec[address]; !ok {
			rf.suspicion = append(rf.suspicion, address)
			rf.cntSuspicion[address] = 1
		}
	}
}


func (rf *BWRaft) sendBroadcastByz(address string, args *RPC.BroadcastByzArgs) (bool, *RPC.BroadcastByzRely) {
	// Detect RPC 中的Client端
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
	client := RPC.NewDetectClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err3 := client.UpdateByzantine(ctx, args)
	if err3 != nil {
		fmt.Println("接收UpdateByzantine结果失败:",err3)
		return false, reply
	}
	return true, reply
}

func (rf *BWRaft) UpdateByzantine(ctx context.Context, args *RPC.BroadcastByzArgs) (*RPC.BroadcastByzRely, error) {
	//返回自己的byz和sus信息
	returnByz, _ := json.Marshal(rf.byzantine)
	returnSus, _ := json.Marshal(rf.suspicion)
	reply := &RPC.BroadcastByzRely{
		ReceiveByzantine:	returnByz,
		ReceiveSuspicion:	returnSus,
	}
	//获得接收的信息
	var receiveByz []string
	var receiveSus []string
	err := json.Unmarshal(args.SendByzantine, &receiveByz)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(args.SendSuspicion, &receiveSus)
	if err != nil {
		fmt.Println(err)
	}
	//根据接收到的信息更新自己的byz信息
	mp := make(map[string]bool)
	for _, add := range rf.byzantine {
		mp[add] = true
	}
	for _, add := range receiveByz {
		if _, ok := mp[add]; !ok {
			mp[add] = true
			rf.byzantine = append(rf.byzantine, add)
		}
	}
	//根据suspicion更新自己的byz信息
	f := len(rf.byzantine)
	for _, add := range receiveSus {
		rf.cntSuspicion[add]++
		if rf.cntSuspicion[add] >= f+1 {
			rf.byzantine = append(rf.byzantine, add)
		}
	}
	return reply, nil
}

//向集群中的其他节点广播自己的拜占庭信息和怀疑的节点的信息，接收其他节点的返回信息，并更新自己的相关信息
func (rf *BWRaft) broadcastByzAndSus() {
	//标记自己已知的拜占庭节点信息
	mp := make(map[string] bool) //标记自己已知的拜占庭节点
	for _, address := range rf.byzantine {
		mp[address] = true
	}

	for _, address := range rf.detectAdd {
		if address == rf.address {
			continue
		}
		go func (add string) {
			byz, _ := json.Marshal(rf.byzantine)
			sus, _ := json.Marshal(rf.suspicion)
			args := &RPC.BroadcastByzArgs{
				SendByzantine:	byz,
				SendSuspicion:	sus,
			}

			fmt.Printf("%s向%s广播byz和sus节点\n", rf.address, add)
			ret, reply := rf.sendBroadcastByz(add+"3", args)
			if ret {
				var receiveByz []string
				var receiveSus []string
				err := json.Unmarshal(reply.ReceiveByzantine, &receiveByz)
				if err != nil {
					fmt.Println(err)
				}
				err = json.Unmarshal(reply.ReceiveSuspicion, &receiveSus)
				if err != nil {
					fmt.Println(err)
				}
				//更新自己已知的拜占庭节点
				for _, add := range receiveByz {
					if _, ok := mp[add]; !ok {
						mp[add] = true
						rf.byzantine = append(rf.byzantine, add)
					}
				}
				//由统计到的suspicion数更新已知的拜占庭节点
				f := len(rf.byzantine)
				for _, add := range receiveSus {
					rf.cntSuspicion[add]++
					if rf.cntSuspicion[add] >= f+1 {
						mp[add] = true
						rf.byzantine = append(rf.byzantine, add)
					}
				}
			} else {
				fmt.Printf("%s未成功接收%s的UpdateByzantine返回消息\n", rf.address, add)
			}
		}(address)
	}
	return
}
