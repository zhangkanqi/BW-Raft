package main

import "fmt"

func main() {
	a := make([]int, 1)
	fmt.Println(len(a), a[len(a)-1])
	a = append(a, 5)
	a = append(a, 6)
	a = append(a, 7)
	n := len(a)
	fmt.Println("all", n)
	for i := 0; i < n; i++ {
		fmt.Println(i, a[i])
	}

}
