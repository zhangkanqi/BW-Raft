package main

import (
	BWRAFT "../Raft"
	PERSISTER "../persist"
	"flag"
	"strings"
	"sync"
	"time"
)

func main() {
	var add = flag.String("address", "", "servers's address")
	var clu = flag.String("cluster", "", "whole cluster's address")
	flag.Parse()
	address := *add
	cluster := strings.Split(*clu, ",")

	persist := &PERSISTER.Persister{}
	persist.Init("../db"+address+time.Now().Format("20060102"))

	BWRAFT.MakeSecretary(address, cluster, persist, &sync.Mutex{})
	time.Sleep(time.Minute * 10)
	//time.Sleep(time.Second*10)
}