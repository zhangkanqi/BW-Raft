package main

import (
	PERSIST "../persist"
	"../testRPC"
	"context"
	"encoding/json"
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

	p := PERSIST.Persister{}
	p.Init("../db102:21:21:45:5000"+time.Now().Format("20060102"))
	po, _ := json.Marshal(p)
	args := &testRPC.KKQQArgs{S: "ss", Pointer:po}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
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