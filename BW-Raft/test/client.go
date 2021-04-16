package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

func main() {
	address := "192.168.8.6:5000"
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
	client := NewKKQQClient(conn)
	args := &KKQQArgs{S:"ss"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reply, err := client.IsKKQQ(ctx, args)
	if err != nil {
		panic(err.Error())
	}
	if reply != nil {
		fmt.Println("接受返回信息成功 ", reply.Success)
	} else {
		fmt.Printf("接受返回结果超时")
	}
}