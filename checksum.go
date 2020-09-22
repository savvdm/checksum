package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var data map[string]string

const separator = "  " // Two-space separator used by sha1sum on Linux

func help() {
	fmt.Println("Specify checsum file name")
}

func check(e error, code int) {
	if e != nil {
		fmt.Println(e)
		os.Exit(code)
	}
}

func readData(file string) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return // input file not found, ok
	}

	f, err := os.Open(file)
	check(err, 3)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.SplitN(scanner.Text(), separator, 2)
		checksum, file := fields[0], fields[1]
		data[file] = checksum // TODO: check for duplicate keys
	}
	err = scanner.Err()
	check(err, 4)
}

func main() {

	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}

	data = make(map[string]string)

	readData(os.Args[1])

	fmt.Printf("Read %d checksums\n", len(data))
}
