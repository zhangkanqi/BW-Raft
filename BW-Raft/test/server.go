package main

import (
	"context"
	"google.golang.org/grpc"
	"net"
)

type AA struct {

}
func (aa *AA) IsKKQQ(ctx context.Context, agrs *KKQQArgs) (*KKQQReply, error) {
	p := agrs.S
	reply := &KKQQReply{Success:true}
	return reply, nil
}
func main() {
	server := grpc.NewServer()
	RegisterKKQQServer(server, new(AA))
	address := "192.168.8.6:5000"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(err.Error())
	}
	err = server.Serve(lis)
	if err != nil {
		panic(err.Error())
	}

}