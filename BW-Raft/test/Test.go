package main

import (
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
	fmt.Println(dbAddress)
}
