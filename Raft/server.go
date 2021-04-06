package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct  {
	address string
	members[] string
	mu *sync.Mutex
	rf *Raft
	persist *Persister
	applyCh chan ApplyMsg

}

func (sv *Server) WriteRequest(ctx context.Context, args *WriteArgs) (*WriteReply, error) {
	reply := &WriteReply{}
	_, reply.IsLeader = sv.rf.getState()
	if !reply.IsLeader {
		return reply, nil
	}
	originalOp := Op{
		key:		args.Key,
		value:		args.Value,
		option: 	"write",
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

func (sv *Server) ReadRequest(ctx context.Context, args *ReadArgs) (*ReadReply, error) {
	reply := &ReadReply{}
	_, reply.IsLeader = sv.rf.getState()
	if !reply.IsLeader {
		return reply, nil
	}
	originalOp := Op{
		option:		"read",
		key:		args.Key,
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
	fmt.Printf("读取到的内容：%s\n", sv.rf.persist.Get(args.Key))
	return reply, nil
}

func (sv *Server) registerServer(address string) {
	// Client和集群成员交互 的Server端
	for {
		server := grpc.NewServer()
		RegisterServeServer(server, sv)
		lis, err1 := net.Listen("tcp", address)
		if err1 != nil {
			log.Fatalln(err1)
		}
		err2 := server.Serve(lis)
		if err2 != nil {
			log.Fatalln(err2)
		}
	}
}

func main() {
	var add = flag.String("address", "", "servers's address")
	var mems = flag.String("members", "", "other members' address")
	flag.Parse()
	address := *add
	members := strings.Split(*mems, ",")
	persist := &Persister{}
	persist.Init("../db"+address)
	sv := Server{
		address: address,
		members: members,
		persist: persist,
	}
	go sv.registerServer(sv.address+"1")
	sv.rf = MakeRaft(sv.address, sv.members, sv.persist, sv.mu)
	time.Sleep(time.Minute * 2)
}