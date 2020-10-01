package main

import (
	"fmt"
)

type statusMap map[string]int

const (
	ADDED    = "Added"
	REPLACED = "Replaced"
	DELETED  = "Deleted"
	CHECKED  = "Checked"
	SKIPPED  = "Skipped"
)

func (status statusMap) register(s string, file string) {
	if v, ok := status[s]; ok {
		status[s] = v + 1
	} else {
		status[s] = 1
	}
	if len(file) > 0 {
		fmt.Println(string(s[0]), file)
	}
}

func (status statusMap) print(keys []string) {
	for _, s := range keys {
		fmt.Printf("%s: %d\n", s, status[s])
	}
}

func (status statusMap) sum(keys []string) (count int) {
	for _, s := range keys {
		count += status[s]
	}
	return
}
