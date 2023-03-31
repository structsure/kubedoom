package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {

	argsWithProg := os.Args
	fmt.Println(argsWithProg)
	argsWithoutProg := os.Args[1:]
	fmt.Println(argsWithoutProg)

	arg := strings.Join(argsWithoutProg, " ")
	fmt.Println(arg)
	fmt.Println(hash(arg))
}

func hash(input string) int32 {
	var hash int32
	hash = 5381
	for _, char := range input {
		hash = ((hash << 5) + hash + int32(char))
	}
	if hash < 0 {
		hash = 0 - hash
	}
	return hash
}
