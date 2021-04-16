package main

import (
	RPC "../RPC"
	BW_RAFT "../Raft"
	PERSISTER "../persist"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"
)

type Secretary struct {
	mu *sync.Mutex
	currentTerm int32
	log []BW_RAFT.Entry
	address string
	cluster[] string //整个集群成员的地址，包括leader和followerß

	commitIndex int32
	lastApplied int32

	nextIndex []int32
	matchIndex []int32

	appendLogCh chan bool
	replicateCh chan bool
	Persist *PERSISTER.Persister
}

func (se *Secretary) getPrevLogIndex(i int) int32 {
	// 每个peer对应的nextIndex会变，其相应的prevLogIndex也会变
	fmt.Printf("NextIndex[%s] = %d, prevLogIndex=%d\n", se.cluster[i], se.nextIndex[i], se.nextIndex[i]-1)
	return se.nextIndex[i]-1 //
}

func (se *Secretary) getPrevLogTerm(i int) int32 {
	prevLogIndex := se.getPrevLogIndex(i)
	fmt.Printf("########prevLogIndex:%d########\n", prevLogIndex)
	fmt.Printf("########prevLogIndexTerm:%d########\n", se.log[prevLogIndex].Term)
	return se.log[prevLogIndex].Term
}

func (se *Secretary) startAppendEntries() {
	fmt.Printf("############Secretary %s--开始日志追加--term:%d ############\n",se.address, se.currentTerm)
	n := len(se.cluster)
	for i := 0; i < n; i++ {
		//随机发送，过半即apply
		rand.Seed(time.Now().UnixNano())
		id := (i + rand.Intn(n)) % n // 不向leader append！！
		go func(idx int) { //
			for {
				appendEntries := se.log[se.nextIndex[idx]:] // 刚开始是空的，为了维持自己的leader身份
				if len(appendEntries) > 0 {
					fmt.Printf("待追加日志长度appendEntries：%d (首)信息：term：%d cm:%s\n", len(appendEntries), appendEntries[0].Term, appendEntries[0].Command)
				}
				entries, _ := json.Marshal(appendEntries) // 转码
				args := &RPC.AppendEntriesArgs{ //prevLogIndex和prevLogTerm会因为冲突而改变，所以放在了for循环内
					Term:          se.currentTerm,
					LeaderId:      0,
					PrevLogIndex:   se.getPrevLogIndex(idx),
					PrevLogTerm:    se.getPrevLogTerm(idx),
					Entries:       entries,
					LeaderCommit:  se.commitIndex,
				}
				if len(appendEntries) == 0 {
					fmt.Printf("%s --> %s  HeartBeat RPC\n", se.address, se.cluster[idx])
				} else {
					fmt.Printf("%s --> %s  AppendEntries RPC\n", se.address, se.cluster[idx])
				}
				ret, reply := se.sendAppendEntries(se.cluster[idx], args)
				if ret {
					if reply.Term > se.currentTerm {
						// se.beFollower(reply.Term)
						return
					}
					if reply.Term != se.currentTerm { //只处理最新term（se.currentTerm）的数据
						return
					}
					if reply.Success {
						se.matchIndex[idx] = args.PrevLogIndex + int32(len(appendEntries)) // 刚开始日志还未复制，matchIndex=prevLogIndex
						se.nextIndex[idx] = se.matchIndex[idx] + 1
						if len(appendEntries) == 0 {
							fmt.Printf("向 %s 发送心跳包成功，nextIndex[%s]=%d\n", se.cluster[idx], se.cluster[idx], se.nextIndex[idx])
						} else {
							fmt.Printf("向 %s 日志追加成功，nextIndex[%s]=%d\n", se.cluster[idx], se.cluster[idx], se.nextIndex[idx])
						}
						se.updateCommitIndex()
						return
					} else {
						se.nextIndex[idx]--
						fmt.Printf("日志不匹配，更新后nextIndex[%s]=%d\n", se.cluster[idx], se.nextIndex[idx])
					}
				}
			}
		}(id)
	}

}

func (se *Secretary) sendAppendEntries(address string, args *RPC.AppendEntriesArgs) (bool, *RPC.AppendEntriesReply) {
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

func (se *Secretary) updateCommitIndex()  { // 只由leader调用
	n := len(se.matchIndex)
	for i := 0; i  < n; i++ {
		fmt.Printf("matchIndex[%s]=%d\n", se.cluster[i], se.matchIndex[i])
	}
	copyMatchIndex := make([]int32, n)
	copy(copyMatchIndex, se.matchIndex)
	sort.Sort(IntSlice(copyMatchIndex))
	N := copyMatchIndex[n / 2] // 过半
	fmt.Printf("过半的commitIndex：%d\n", N)
	if N > se.commitIndex && se.log[N].Term == se.currentTerm {
		se.commitIndex = N
		fmt.Printf("new LeaderCommit：%d\n", se.commitIndex)
		se.updateLastApplied()
	}
}

func (se *Secretary) updateLastApplied()  { // apply
	fmt.Printf("updateLastApplied()---lastApplied: %d, commitIndex: %d\n", se.lastApplied, se.commitIndex)
	for se.lastApplied < se.commitIndex { // 0 0
		se.lastApplied++
		curEntry := se.log[se.lastApplied]
		cm := curEntry.Command
		if cm.Option == "write" {
			se.Persist.Put(cm.Key, cm.Value)
			fmt.Printf("Write 命令：%s-%s 被apply\n", cm.Key, cm.Value)
		} else if cm.Option == "read" {
			fmt.Printf("Read 命令：key:%s 被apply\n", cm.Key)
		}
		/*
		applyMsg := ApplyMsg{
			true,
			curEntry.Command,
			int(se.lastApplied),
		}
		applyRequest(se.applyCh, applyMsg)
		 */
	}
}

func (se *Secretary) MakeSecretary() {

}


func main() {
	var add = flag.String("address", "", "servers's address")
	var clu = flag.String("cluster", "", "whole cluster's address")
	flag.Parse()
	address := *add
	cluster := strings.Split(*clu, ",")
	se := &Secretary{
		address: address,
		cluster: cluster,
	}

}