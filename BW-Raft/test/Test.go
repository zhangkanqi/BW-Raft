package main

import (
	PERSISTER "../persist"
	"encoding/json"
	"fmt"
	"time"
)

type Va struct {
	a int
	ch chan bool
}

func (v *Va) add() {
	v.a++
	go func() {
		v.a++
		if v.a == 200 {
			fmt.Printf("200~~~~\n")
		}
	}()

}

func main() {

	// go 套 go
	address := "102:21:21:45:5000"
	dbAddress := "db"+address+time.Now().Format("20060102")
	fmt.Println(dbAddress)

	va := &Va{a:0}
	va.ch = make(chan bool, 1)
	va.ch <- true
	select {
	case <- va.ch:
		fmt.Printf("ch 中有东西\n")
		select {
		case <- va.ch:
			fmt.Printf("ch 中还有东西\n")
		default:
			fmt.Printf("ch 中的东西被清空了\n")
		}
	default:
		fmt.Printf("ch 中没有东西\n")
	}

	for i := 0; i < 200; i++ {
		go va.add()
	}

	go func() {
		fmt.Println("gogogogog")
	}()

	a := make([]int32, 10)
	c := make([]int32, 10)
	a[1] = 3
	b, _ :=json.Marshal(a)
	err := json.Unmarshal(b, &c)
	if err != nil {
		fmt.Println(err)
	}
	for i, j := range c {
		fmt.Println(i, j)
	}
	/*
	p := &PERSISTER.Persister{}
	p.Init("../db"+address+time.Now().Format("20060102"))
	p.Put("1", "1")
	p.Put("2", "2")
	p.Put("3", "3")
	fmt.Printf("p: key:%s-value:%s\n", "2", p.Get("2"))
	//一次只允许一个进程访问一个特定的数据库
	p.Close()
	*/

	q := &PERSISTER.Persister{}
	q.Init("../db"+address+time.Now().Format("20060102"))
	fmt.Printf("q: key:%s-value:%s\n", "2", q.Get("2"))


}
