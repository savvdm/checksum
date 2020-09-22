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

func parseLine(line string) (checksum, file string) {
	fields := strings.SplitN(line, separator, 2)
	if len(fields) != 2 {
		fmt.Printf("Invalid input line: '%s'\n", line)
		os.Exit(5)
	}

	checksum, file = fields[0], fields[1]
	if len(checksum) == 0 {
		fmt.Printf("Invalid input - checksum is missing: '%s'\n", line)
		os.Exit(5)
	}
	if len(file) == 0 {
		fmt.Printf("Invalid input - file path is missing: '%s'\n", line)
		os.Exit(5)
	}
	return
}

func readData(file string) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		fmt.Println("File not found:", file)
		return // input file not found, ok
	}

	f, err := os.Open(file)
	check(err, 3)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		checksum, file := parseLine(scanner.Text())
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
