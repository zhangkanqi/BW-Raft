package main

import (
	RPC "../RPC"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Client struct {
	cluster[] string
	leaderId int
}

var count int32 = 0

func (ct *Client) sendWriteRequest(address string, args *RPC.WriteArgs) (bool, *RPC.WriteReply) {
	// WriteRequest 的 Client端 拨号
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		panic(err1)
	}
	defer func () {
		err2 := conn.Close()
		if err2 != nil {
			panic(err1)
		}
	}()
	client := RPC.NewServeClient(conn)
	reply, err3 := client.WriteRequest(context.Background(), args)
	if err3 != nil {
		fmt.Println(err3)
		return false, reply
	}
	return true, reply
}

func (ct *Client) sendReadRequest(address string, args *RPC.ReadArgs) (bool, *RPC.ReadReply) {
	// ReadRequest 的 Client端， 拨号
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		panic(err1)
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			panic(err2)
		}
	}()
	client := RPC.NewServeClient(conn)
	reply, err3 := client.ReadRequest(context.Background(), args)
	if err3 != nil {
		fmt.Println(err3)
		return false, reply
	}
	return true, reply
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
		ret, reply := ct.sendWriteRequest(ct.cluster[id], args)
		if ret {
			if !reply.IsLeader {
				//fmt.Printf("Write请求，%s 不是Leader, id++\n", ct.cluster[id])
				id = (id + 1) % n
			} else {
				break //找到了leaderId，结束for
			}
		} else {
			fmt.Printf("send WriteRequest 返回false\n")
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
		ret, reply := ct.sendReadRequest(ct.cluster[id], args)
		if ret {
			if !reply.IsLeader {
				//fmt.Printf("Read请求，%s 不是Leader, id++\n", ct.cluster[id])
				id = (id + 1) % n
			} else {
				break
			}
		} else {
			fmt.Printf("send ReadRequest 返回false\n")
		}
	}
	ct.leaderId = id
}

func (ct *Client) startWriteRequest(num int) {
	var key, value string
	for i := 0; i < num; i++ {
		key = strconv.Itoa(i+1)
		value = strconv.Itoa(i+1)
		fmt.Printf("	·······start write %s-%s········\n", key, value)
		ct.Write(key, value)
		//等待上一个write命令运行完，否则并行执行appendEntries时，可能会出现数组越界的情况
		atomic.AddInt32(&count, 1)
		//time.Sleep(time.Millisecond*10)
	}
}

func (ct *Client) startReadRequest(num int) {
	var key string
	for i := 0; i < num; i++ {
		key = strconv.Itoa(i+1)
		fmt.Printf("	·······start read key：%s········\n", key)
		ct.Read(key)
		atomic.AddInt32(&count, 1)
		//time.Sleep(time.Millisecond*50)
	}
}

func end(clientNum, operationNum int32) {
	start :=  time.Now().UnixNano()
	for {
		if count == clientNum * operationNum {
			end :=  time.Now().UnixNano()
			fmt.Printf("%d个请求处理完成，用时%dns\n", clientNum * operationNum, end-start)
			fmt.Printf("平均每个操作用时%dns\n", (end-start)/int64(clientNum * operationNum))
			return
		}
	}
}

func main() {
	var md = flag.String("mode", "", "all observers")
	var cnum = flag.String("clientNum", "", "the number of clients making requests in parallel")
	var onum = flag.String("operationNum", "", "the number of requests made by per client")
	var clu = flag.String("cluster", "", "all cluster members' address")
	flag.Parse()
	mode := *md
	clientNum, _ := strconv.Atoi(*cnum)
	operationNum, _ := strconv.Atoi(*onum)
	clientNum32 := int32(clientNum)
	operationNum32 := int32(operationNum)
	cluster := strings.Split(*clu, ",")

	fmt.Println("集群成员：")
	n := len(cluster)
	for i := 0; i < n; i++ {
		cluster[i] = cluster[i] + "1"
		fmt.Println(cluster[i])
	}

	ct := Client{}
	ct.leaderId = 0
	ct.cluster = cluster

	go end(clientNum32, operationNum32)

	if mode == "write" {
		for i := 0; i < clientNum; i++ {
			go ct.startWriteRequest(operationNum)
		}
	}
	if mode == "read" {
		for i := 0; i < clientNum; i++ {
			go 	ct.startReadRequest(operationNum)
		}
	}
	fmt.Println("每秒处理的请求数：")
	time.Sleep(time.Second * 3)
	fmt.Println(count / 3)

	time.Sleep(time.Minute * 5)
}