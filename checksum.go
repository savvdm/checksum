package main

import (
	"fmt"
	"os"
	"time"
)

var data map[string]string

func help() {
	fmt.Println("Specify checsum file name")
}

func readData(file string) (time.Time, error) {
	return time.Now(), nil
}

func main() {

	if len(os.Args) < 2 {
		help()
		os.Exit(1)
	}

	data = make(map[string]string)

	var checksumTime time.Time
	if checksumTime, err := readData(os.Args[1]); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

}
