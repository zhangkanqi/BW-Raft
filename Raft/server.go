package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct  {
	address string
	members[] string
	mu *sync.Mutex
	rf *Raft
	persist *Persister
	applyCh chan ApplyMsg

}

func (sv *Server) WriteRequest(ctx context.Context, args *WriteArgs) (*WriteReply, error) {

}

func (sv *Server) ReadRequest(ctx context.Context, args *ReadArgs) (*ReadReply, error) {

}

func (sv *Server) registerServer(address string) {
	// Client和集群成员交互 的Server端
	for {
		server := grpc.NewServer()
		RegisterServeServer(server, sv)
		lis, err1 := net.Listen("tcp", address)
		if err1 != nil {
			log.Fatalln(err1)
		}
		err2 := server.Serve(lis)
		if err2 != nil {
			log.Fatalln(err2)
		}
	}
}

func main() {
	var add = flag.String("address", "", "servers's address")
	var mems = flag.String("members", "", "other members' address")
	flag.Parse()
	address := *add
	members := strings.Split(*mems, ",")
	persist := &Persister{}
	persist.Init("../db"+address)
	sv := Server{
		address: address,
		members: members,
		persist: persist,
	}
	go sv.registerServer(sv.address+"1")
	sv.rf = MakeRaft(sv.address, sv.members, sv.persist, sv.mu)
	time.Sleep(time.Minute * 2)
}