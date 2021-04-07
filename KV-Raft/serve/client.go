package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"strings"
	RPC "../RaftRPC"
)

type Client struct {
	cluster[] string
	leaderId int
}

func (ct *Client) sendWriteRequest(address string, args *RPC.WriteArgs) *RPC.WriteReply {
	// WriteRequest 的 Client端 拨号
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		log.Fatalln(err1)
	}
	defer func () {
		err2 := conn.Close()
		if err2 != nil {
			log.Fatalln(err2)
		}
	}()
	client := RPC.NewServeClient(conn)
	reply, err3 := client.WriteRequest(context.Background(), args)
	if err3 != nil {
		log.Fatalln(err3)
	}
	return reply
}

func (ct *Client) sendReadRequest(address string, args *RPC.ReadArgs) *RPC.ReadReply {
	// ReadRequest 的 Client端， 拨号
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		log.Fatalln(err1)
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			log.Fatalln(err2)
		}
	}()
	client := RPC.NewServeClient(conn)
	reply, err3 := client.ReadRequest(context.Background(), args)
	if err3 != nil {
		log.Fatalln(err3)
	}
	return reply
}

func (ct *Client) Write(key, value string) {
	// 重定向到leader
	args := &RPC.WriteArgs{
		Key:           key,
		Value:         value,
	}
	id := ct.leaderId
	n := len(ct.cluster)
	for {
		reply := ct.sendWriteRequest(ct.cluster[id], args)
		if !reply.IsLeader {
			fmt.Printf("Write请求，%s 不是Leader, id++\n", ct.cluster[id])
			id = (id + 1) % n
		} else {
			break
		}
	}
	ct.leaderId = id
}

func (ct *Client) Read(key string) {
	//重定向到leader
	args := &RPC.ReadArgs{Key:key}
	id := ct.leaderId
	n := len(ct.cluster)
	for {
		reply := ct.sendReadRequest(ct.cluster[id], args)
		if !reply.IsLeader {
			fmt.Printf("Read请求，%s 不是Leader, id++\n", ct.cluster[id])
			id = (id + 1) % n
		} else {
			break
		}
	}
	ct.leaderId = id
}

func (ct *Client) startWriteRequest() {
	n := len(ct.cluster)
	for i := 0; i < n; i++ {
		ct.cluster[i] = ct.cluster[i] + "1"
	}
	num := 5 // 最简单5个write 5个read
	var key, value string
	for i := 0; i < num; i++ {
		key = strconv.Itoa(i+1)
		value = strconv.Itoa(i+1)
		ct.Write(key, value)
	}
}

func (ct *Client) startReadRequest() {
	n := len(ct.cluster)
	for i := 0; i < n; i++ {
		ct.cluster[i] = ct.cluster[i] + "1" // ？
	}
	num := 5 // 最简单5个write 5个read
	var key string
	for i := 0; i < num; i++ {
		key = strconv.Itoa(i+1)
		ct.Read(key)
	}
}

func main() {
	var clu = flag.String("cluster", "", "all cluster members' address")
	flag.Parse()
	cluster := strings.Split(*clu, ",")
	fmt.Println("集群成员：")
	n := len(cluster)
	for i := 0; i < n; i++ {
		fmt.Println(cluster[i])
	}
	ct := Client{}
	ct.leaderId = 0
	ct.cluster = cluster
	ct.startWriteRequest()
	ct.startReadRequest()
}