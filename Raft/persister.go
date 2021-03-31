package Raft

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

type Persister struct {
	path string
	db *leveldb.DB
}

func (p *Persister) Init(path string) {
	var err error
	p.db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (p *Persister) Put(key string, value string) {
	err := p.db.Put([]byte(key), []byte(value), nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (p *Persister) Get(key string) []byte {
	value, err := p.db.Get([]byte(key), nil)
	if err != nil {
		log.Fatalln(err)
	}
	return value
}

func (p *Persister) PrintStrVal(key string) {
	value := p.Get(key)
	fmt.Println(string(value))
}