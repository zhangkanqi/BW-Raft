package main

import (
	RPC "../RPC"
	RAFT "../Raft"
	PERSISTER "../persist"
	"flag"
	//"fmt"
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

}

func (sv *Server) WriteRequest(ctx context.Context, args *RPC.WriteArgs) (*RPC.WriteReply, error) {

	//fmt.Printf("\n·····1····进入%s的WriteRequest处理端·········\n", sv.address)
	reply := &RPC.WriteReply{}
	_, reply.IsLeader = sv.rf.GetState()
	if !reply.IsLeader {
		//fmt.Printf("\n·········%s 不是leader，return false·········\n", sv.address)
		return reply, nil
	}
	request := RAFT.Op{
		Key:    args.Key,
		Value:  args.Value,
		Option: "write",
	}
	//index, _, isLeader := sv.rf.Start(request)
	_, _, isLeader := sv.rf.Start(request)
	if !isLeader {
		//fmt.Printf("******* %s When write, Leader change!*****\n", sv.address)
		reply.IsLeader = false
		return reply, nil
	}
	//apply := <- sv.applyCh
	//fmt.Printf("Server端--新指令的内容：%s %s %s, 新指令的index：%d\n", request.Option, request.Key, request.Value, index)
	reply.IsLeader = true
	reply.Success = true
	//fmt.Printf("·····2····%s的WriteRequest处理成功·········\n", sv.address)

	return reply, nil
}

func (sv *Server) ReadRequest(ctx context.Context, args *RPC.ReadArgs) (*RPC.ReadReply, error) {
	//fmt.Printf("\n·····1····进入%s的ReadRequest处理端·········\n", sv.address)
	reply := &RPC.ReadReply{}
	_, reply.IsLeader = sv.rf.GetState()
	if !reply.IsLeader {
		//fmt.Printf("\n·········%s 不是leader，return false·········\n", sv.address)
		return reply, nil
	}
	request := RAFT.Op{
		Option:		"read",
		Key:		args.Key,
	}
	//index, _, isLeader := sv.rf.Start(request)
	_, _, isLeader := sv.rf.Start(request)
	if !isLeader {
		//fmt.Printf("******* %s When read, Leader change!*****\n", sv.address)
		reply.IsLeader = false
		return reply, nil
	}
	reply.IsLeader = true
	//apply := <- sv.applyCh
	//fmt.Printf("新指令的内容：%s %s, 新指令的index：%d\n", request.Option, request.Key, index)
	//读取的内容
	//fmt.Printf("读取到的内容：%s\n", sv.rf.Persist.Get(args.Key))
	//fmt.Printf("·····2····%s的ReadRequest处理成功·········\n", sv.address)
	return reply, nil
}


func (sv *Server) registerServer(address string) {
	// Client和集群成员交互 的Server端
	//fmt.Println("················进入外部注册服务器········")
	lis, err1 := net.Listen("tcp", address)
	if err1 != nil {
		//fmt.Println(err1)
	}
	server := grpc.NewServer()
	RPC.RegisterServeServer(server, sv)
	err2 := server.Serve(lis)
	if err2 != nil {
		//fmt.Println(err2)
	}
}


func main() {
	var add = flag.String("address", "", "servers's address")
	var mems = flag.String("members", "", "other members' address")
	flag.Parse()
	address := *add
	members := strings.Split(*mems, ",")
	persist := &PERSISTER.Persister{}

	persist.Init("../db"+address+string(time.Now().Unix()))
	sv := Server{
		address: address,
		members: members,
		persist: persist,
	}
	go sv.registerServer(sv.address+"1")
	sv.rf = RAFT.MakeRaft(sv.address, sv.members, sv.persist, sv.mu)
	time.Sleep(time.Minute * 2)
}

