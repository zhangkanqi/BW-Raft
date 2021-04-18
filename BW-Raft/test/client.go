package main

import (
	"../testRPC"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

func main() {
	t1 := time.Now().UnixNano()
	//address := "192.168.8.6:5000"
	address := "localhost:8090"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		err2 := conn.Close()
		if err2 != nil {
			fmt.Println(err2)
		}
	}()
	client := testRPC.NewKKQQClient(conn)

	args := &testRPC.KKQQArgs{S: "ss"}
	// 2ms
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	reply, err := client.IsKKQQ(ctx, args)
	t2 := time.Now().UnixNano()
	if reply != nil {
		fmt.Println("接受返回信息成功 ", reply.Success, t2-t1)
	} else {
		fmt.Printf("接受返回结果超时\n")
	}
	if err != nil {
		panic(err.Error())
	}

}