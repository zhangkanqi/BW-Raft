package main

import (
	RPC "../RPC"
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

func (ob *Observer) WriteRequest(ctx context.Context, args *RPC.WriteArgs) (*RPC.WriteReply, error) {
	reply := &RPC.WriteReply{}
	return reply, nil
}

func (ob *Observer) ReadRequest(ctx context.Context, readArgs *RPC.ReadArgs) (*RPC.ReadReply, error) {
	fmt.Printf("\n·····1····进入observer %s 的ReadRequest处理端·········\n", ob.address)
	readReply := &RPC.ReadReply{}
	n := len(ob.cluster)
	for i := 0; i < n; i++ {
		arg := &RPC.GetValueArgs{Key:readArgs.Key} // 50002端口
		success, getValueReply := ob.judgeConnect(ob.cluster[i], arg)
		if success { //未连接超时
			fmt.Printf("\nObserver-%s 和 Cluster-%s 连接成功，返回结果%s-%s·········\n\n", ob.address, ob.cluster[i], readArgs.Key, getValueReply.Value)
			readReply.Value = getValueReply.Value
			return readReply, nil
		}
	}
	return readReply, nil
}

func (ob *Observer) judgeConnect(address string, args *RPC.GetValueArgs) (bool, *RPC.GetValueReply) {
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
	client := RPC.NewGetValueClient(conn)
	//正常连接时2ms即可返回结果
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	reply, err3 := client.GetValue(ctx, args)
	if err3 != nil {
		fmt.Println("接受Connect结果失败:",err3)
		return false, reply
	}
	return true, reply
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
	ob.registerServer(address+"1")
	time.Sleep(time.Minute * 2)
}
