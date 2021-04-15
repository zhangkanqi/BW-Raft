package main

import (
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"sort"
)

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
