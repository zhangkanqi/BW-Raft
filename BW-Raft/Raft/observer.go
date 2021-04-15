package main

import (
	RPC "../RPC"
	PERSISTER "../persist"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"strings"
	"time"
)

type Observer struct {
	address string
	//处理read请求时直接从某个能正常连接的Follower的persist中去读
	cluster[] string
}

func (ob *Observer) ReadRequest(ctx context.Context, args *RPC.ReadArgs) (*RPC.ReadReply, error) {
	fmt.Printf("\n·····1····进入observer %s 的ReadRequest处理端·········\n", ob.address)
	reply := &RPC.ReadReply{}
	n := len(ob.cluster)
	for i := 0; i < n; i++ {
		arg := &RPC.ConnectArgs{}
		if ob.judgeConnect(ob.cluster[i], arg) {
			p := &PERSISTER.Persister{}
			p.Init("../db"+ob.cluster[i]+time.Now().Format("20060102"))
			reply.Value = p.Get(args.Key)
			fmt.Printf("·····2····observer %s 连接%s成功，返回结果%s-%s·········\n\n", ob.address, ob.cluster[i], args.Key, reply.Value)
			return reply, nil
		}
	}
	return reply, nil
}

func (ob *Observer) judgeConnect(address string, args *RPC.ConnectArgs) bool {
	// Connect RPC 的client端
	address = address+"2" //50002
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
	client := RPC.NewConnectClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	reply, err3 := client.IsConnect(ctx, args)
	if err3 != nil {
		fmt.Println("接受Connect结果失败:",err3)
		return false
	}
	return reply.Success
}

func (ob *Observer) registerServer(address string) {
	fmt.Println("················observer注册服务器········")
	lis, err1 := net.Listen("tcp", address)
	if err1 != nil {
		fmt.Println(err1)
	}
	server := grpc.NewServer()
	RPC.RegisterServeServer(server, ob)
	err2 := server.Serve(lis)
	if err2 != nil {
		fmt.Println(err2)
	}
}

func main() {
	var add = flag.String("address", "", "observer's address")
	var clu = flag.String("cluster", "", "whole cluster's address")
	flag.Parse()
	address := *add
	cluster := strings.Split(*clu, ",")
	ob := &Observer{
		address: address,
		cluster: cluster,
	}
	ob.registerServer(address)
	time.Sleep(time.Minute * 2)
}
