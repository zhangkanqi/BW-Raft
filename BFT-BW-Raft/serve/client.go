package main

import (
	"../RPC"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	cluster[] string
	sectary[] string
	readCluster[] string
	leaderId int
}

// 50001 WriteRequest
func (ct *Client) sendWriteRequest(address string, args *RPC.WriteArgs) (bool, *RPC.WriteReply) {
	// WriteRequest 的 Client端 拨号
	conn, err1 := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err1 != nil {
		fmt.Println(err1)
	}
	defer func () {
		err2 := conn.Close()
		if err2 != nil {
			fmt.Println(err2)
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
				fmt.Printf("Write请求，%s 不是Leader, id++\n", ct.cluster[id])
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

//50001 ReadRequest
func (ct *Client) sendReadRequest(address string, args *RPC.ReadArgs) (bool, *RPC.ReadReply) {
	// ReadRequest 的 Client端， 拨号
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
	client := RPC.NewServeClient(conn)
	reply, err3 := client.ReadRequest(context.Background(), args)
	if err3 != nil {
		fmt.Println(err3)
		return false, reply
	}
	return true, reply
}

func (ct *Client) Read(key string) {
	//重定向到leader
	args := &RPC.ReadArgs{Key:key}
	n := len(ct.readCluster) //cluster+observers
	for i := 0; i < n; i++ {
		rand.Seed(time.Now().UnixNano())
		id := (i + rand.Intn(n)) % n
		fmt.Printf("Client send ReadRequest to %s\n", ct.readCluster[id])
		ret, reply := ct.sendReadRequest(ct.readCluster[id], args) // 50001端口
		if ret {
			fmt.Printf("读取到的内容：%s-%s\n", args.Key, reply.Value)
			return
		}
	}
}

func (ct *Client) startWriteRequest() {
	num := 15 // 最简单5个write 5个read
	var key, value string
	for i := 0; i < num; i++ {
		key = strconv.Itoa(i+1)
		value = strconv.Itoa(i+1)
		fmt.Printf("	·······start write %s-%s········\n", key, value)
		ct.Write(key, value)
		//等待上一个write命令运行完，否则并行执行appendEntries时，可能会出现数组越界的情况
		time.Sleep(time.Millisecond*10)
	}
}

func (ct *Client) startReadRequest() {
	num := 15 // 最简单5个write 5个read
	var key string
	for i := 0; i < num; i++ {
		key = strconv.Itoa(i+1)
		fmt.Printf("	·······start read key：%s········\n", key)
		ct.Read(key)
		//time.Sleep(time.Millisecond*50)
	}
}


func main() {
	var clu = flag.String("cluster", "", "all cluster members' address")
	var ob = flag.String("observers", "", "all observers")
	flag.Parse()
	cluster := strings.Split(*clu, ",")
	observers := strings.Split(*ob, ",")

	fmt.Println("集群成员：")
	n := len(cluster)
	for i := 0; i < n; i++ {
		cluster[i] = cluster[i] + "1"
		fmt.Println(cluster[i])
	}

	fmt.Println("Observers成员：")
	n = len(observers)
	for i := 0; i < n; i++ {
		observers[i] = observers[i] + "1"
		fmt.Println(cluster[i])
	}

	ct := Client{}
	ct.leaderId = 0
	ct.cluster = cluster
	ct.readCluster = cluster
	ct.readCluster = append(ct.readCluster, observers...)

	ct.startWriteRequest()
	//time.Sleep(time.Millisecond*1000)
	ct.startReadRequest()
}
