package main

import (
	RPC "../RPC"
	BW_RAFT "../Raft"
	PERSISTER "../persist"
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
	secretaries[] string
	observers[] string
	mu *sync.Mutex
	rf *BW_RAFT.BWRaft
	persist *PERSISTER.Persister

}

func (sv *Server) WriteRequest(ctx context.Context, args *RPC.WriteArgs) (*RPC.WriteReply, error) {

	fmt.Printf("\n·····1····进入cluster-%s 的WriteRequest处理端·········\n", sv.address)
	reply := &RPC.WriteReply{}
	_, reply.IsLeader = sv.rf.GetState()
	if !reply.IsLeader {
		fmt.Printf("\n·········%s 不是leader，return false·········\n", sv.address)
		return reply, nil
	}
	request := BW_RAFT.Op{
		Key:    args.Key,
		Value:  args.Value,
		Option: "write",
	}
	index, _, isLeader := sv.rf.Start(request)
	if !isLeader {
		fmt.Printf("******* %s When write, Leader change!*****\n", sv.address)
		reply.IsLeader = false
		return reply, nil
	}
	//apply := <- sv.applyCh
	fmt.Printf("Server端--新指令的内容：%s %s %s, 新指令的index：%d\n", request.Option, request.Key, request.Value, index)
	reply.IsLeader = true
	reply.Success = true
	fmt.Printf("·····2····%s的WriteRequest处理成功·········\n\n", sv.address)

	return reply, nil
}

// Leader/Follower收到read请求时，直接处理，不进行日志复制
func (sv *Server) ReadRequest(ctx context.Context, args *RPC.ReadArgs) (*RPC.ReadReply, error) {
	fmt.Printf("\n·····1····进入cluster-%s的ReadRequest处理端·········\n", sv.address)
	reply := &RPC.ReadReply{}
	//读取的内容
	fmt.Printf("读取到的内容：%s\n", sv.rf.Persist.Get(args.Key))
	fmt.Printf("·····2····%s的ReadRequest处理成功·········\n\n", sv.address)
	return reply, nil
}


//50001
func (sv *Server) registerServer(address string) {
	// Client和集群成员交互 的Server端
	fmt.Println("················集群成员注册服务器········")
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


func main() {
	var add = flag.String("address", "", "servers's address")
	var mems = flag.String("members", "", "other members' address")
	var sec = flag.String("secretaries", "", "all secretaries")
	var ob = flag.String("observers", "", "all observers")
	flag.Parse()
	address := *add
	members := strings.Split(*mems, ",")
	secretaries := strings.Split(*sec, ",")
	observers := strings.Split(*ob, ",")
	persist := &PERSISTER.Persister{}
	//persist.Init("../db"+address+string(time.Now().Unix()))
	dbAddress := "db"+address+time.Now().Format("20060102")
	fmt.Printf("~~~~~新建存储地址：%s\n", dbAddress)
	persist.Init("../db"+address+time.Now().Format("20060102"))
	//persist.Init(dbAddress)

	sv := Server{
		address: 		address,
		members: 		members,
		secretaries:	secretaries,
		observers:		observers,
		persist: 		persist,
	}

	go sv.registerServer(sv.address+"1")
	sv.rf = BW_RAFT.MakeBWRaft(sv.address, sv.members, sv.secretaries, sv.persist, sv.mu) // 一直不会返回结果
	time.Sleep(time.Minute * 2)
}

