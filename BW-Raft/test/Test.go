package main

import (
	"fmt"
	"time"
)

func main() {
	address := "102:21:21:45:5000"
	dbAddress := "db"+address+time.Now().Format("20060102")
	fmt.Println(dbAddress)
}
