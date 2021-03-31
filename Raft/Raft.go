package Raft

import "sync"

type State int
const NULL int32 = -1

const (
	Follower State = iota // Follower = 0
	Candidate
	Leader
)

type Op struct {
	option string
	key string
	value string
	id int32
	seq int32
}

type Log struct {
	term int32
	Command Op
}

type Raft struct {
	mu *sync.Mutex
	me int32
	address string
	members []string
	role State

	currentTerm int32
	votedFor int32
	log []Log

	commitIndex int32
	lastApplied int32

	nextIndex []int32
	matchIndex []int32

	voteCh chan bool
	appendLogCh chan bool


}



