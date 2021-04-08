package main

import (
	//RAFT "../Raft"
	RPC "../RaftRPC"
	PERSISTER "../persist"
	//TESTRAFT "../Test"
	RAFT "../Test"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct  {
	address string
	members[] string
	mu *sync.Mutex
	rf *RAFT.Raft
	persist *PERSISTER.Persister
	applyCh chan RAFT.ApplyMsg

}
//go run . -address 192.168.8.3:5000 -members 192.168.8.3:5000192.168.8.6:5000,192.168.8.7:5000

func (sv *Server) WriteRequest(ctx context.Context, args *RPC.WriteArgs) (*RPC.WriteReply, error) {
	reply := &RPC.WriteReply{}
	_, reply.IsLeader = sv.rf.GetState()
	if !reply.IsLeader {
		return reply, nil
	}
	originalOp := RAFT.Op{
		Key:    args.Key,
		Value:  args.Value,
		Option: "write",
	}
	index, _, isLeader := sv.rf.Start(originalOp)
	if !isLeader {
		fmt.Println("When write, Leader change!")
		reply.IsLeader = false
		return reply, nil
	}
	apply := <- sv.applyCh
	fmt.Printf("新指令的内容：%s, 新指令的index：%d\n", apply.Command, index)
	reply.Success = true
	return reply, nil
}

func (sv *Server) ReadRequest(ctx context.Context, args *RPC.ReadArgs) (*RPC.ReadReply, error) {
	reply := &RPC.ReadReply{}
	_, reply.IsLeader = sv.rf.GetState()
	if !reply.IsLeader {
		return reply, nil
	}
	originalOp := RAFT.Op{
		Option:		"read",
		Key:		args.Key,
	}
	index, _, isLeader := sv.rf.Start(originalOp)
	if !isLeader {
		fmt.Println("When read, Leader change!")
		reply.IsLeader = false
		return reply, nil
	}
	apply := <- sv.applyCh
	fmt.Printf("新指令的内容：%s, 新指令的index：%d\n", apply.Command, index)
	//读取的内容
	fmt.Printf("读取到的内容：%s\n", sv.rf.Persist.Get(args.Key))
	return reply, nil
}

func (sv *Server) registerServer(address string) {
	// Client和集群成员交互 的Server端
	for {
		fmt.Println("················进入外部注册服务器········")
		lis, err1 := net.Listen("tcp", address)
		if err1 != nil {
			fmt.Println(err1)
		}
		server := grpc.NewServer()
		RPC.RegisterServeServer(server, sv)
		err2 := server.Serve(lis)
		if err2 != nil {
			fmt.Println(err2)
		}
	}
}

func main() {
	var add = flag.String("address", "", "servers's address")
	var mems = flag.String("members", "", "other members' address")
	flag.Parse()
	address := *add
	members := strings.Split(*mems, ",")
	persist := &PERSISTER.Persister{}
	persist.Init("../db"+address)
	sv := Server{
		address: address,
		members: members,
		persist: persist,
	}
	go sv.registerServer(sv.address+"1")
	//sv.rf = RAFT.MakeRaft(sv.address, sv.members, sv.persist, sv.mu)
	sv.rf = RAFT.MakeRaft(sv.address, sv.members, sv.persist, sv.mu)
	time.Sleep(time.Minute * 2)
}