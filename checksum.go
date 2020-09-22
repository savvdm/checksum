package main

import (
	"bufio"
	"fmt"
	"os"
)

var data map[string]string

func help() {
	fmt.Println("Specify checsum file name")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readData(file string) {
	f, err := os.Open(file)
	check(err)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
	err = scanner.Err()
	check(err)
}

func main() {

	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}

	data = make(map[string]string)

	readData(os.Args[1])
}
