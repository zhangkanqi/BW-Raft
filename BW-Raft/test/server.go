package main

import (
	PERSIST "../persist"
	"../testRPC"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type AA struct {

}
func (aa *AA) IsKKQQ(ctx context.Context, agrs *testRPC.KKQQArgs) (*testRPC.KKQQReply, error) {
	p := &PERSIST.Persister{}
	err := json.Unmarshal(agrs.Pointer, p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("q: key:%s-value:%s\n", "2", p.Get("2"))
	reply := &testRPC.KKQQReply{}
	return reply, nil
}
func main() {
	server := grpc.NewServer()
	testRPC.RegisterKKQQServer(server, new(AA))
	//address := "192.168.8.6:5000"
	address := "localhost:8090"
	lis, err := net.Listen("tcp", address)
	//lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err.Error())
	}
	err = server.Serve(lis)
	if err != nil {
		panic(err.Error())
	}

}