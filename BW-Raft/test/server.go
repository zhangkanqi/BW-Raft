package main

import (
	 "../AK"
	"context"
	"google.golang.org/grpc"
	"net"
)

type AA struct {

}
func (aa *AA) IsKKQQ(ctx context.Context, agrs *AK.KKQQArgs) (*AK.KKQQReply, error) {
	p := agrs.S
	reply := &AK.KKQQReply{Success:true}
	return reply, nil
}
func main() {
	server := grpc.NewServer()
	AK.RegisterKKQQServer(server, new(AA))
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