package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Println("j8s CLI: argument received -", os.Args[1])
	} else {
		fmt.Println("j8s CLI: no arguments provided")
	}
}
