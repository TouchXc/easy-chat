package main

import "fmt"

func main() {
	a := 7
	fmt.Printf("%b %v \n", a, a)
	b := a << 2
	fmt.Printf("%b %v \n", b, b)
	c := a >> 2
	fmt.Printf("%b %v \n", c, c)
}
