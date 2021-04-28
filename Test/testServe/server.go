package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	RAFT "../testRaft"
	PERSISTER "../persister"
)

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *RAFT.Raft
	applyCh chan int
	//applyCh chan raft.ApplyMsg
	dead         int32 // set by Kill()
	maxraftstate int   // snapshot if log grows this big
	delay        int
	// Your definitions here.
	Seq     map[int64]int64
	db      map[string]string
	chMap   map[int]chan RAFT.Op
	persist *PERSISTER.Persister
}


func main() {

	var add = flag.String("address", "", "Input Your address")
	var mems = flag.String("members", "", "Input Your follower")
	var delays = flag.String("delay", "", "Input Your follower")

	flag.Parse()

	server := KVServer{}
	// Local address
	address := *add
	persist := &PERSISTER.Persister{}
	persist.Init("../db/" + address)
	//for i := 0; i <= int (address[ len(address) - 1] - '0'); i++{
	server.applyCh = make(chan int, 100)

	//server.applyCh = make(chan int, 1)
	fmt.Println(server.applyCh)

	server.persist = persist

	// Members's address
	members := strings.Split(*mems, ",")
	delay, _ := strconv.Atoi(*delays)
	server.delay = delay
	if delay == 0 {
		fmt.Println("##########################################")
		fmt.Println("### Don't forget input delay's value ! ###")
		fmt.Println("##########################################")

	}
	fmt.Println(address, members, delay)

	//go server.RegisterServer(address + "1")

	server.rf = RAFT.MakeRaft(address, members, persist, &server.mu, server.applyCh, delay)

	time.Sleep(time.Minute * 2)

}
