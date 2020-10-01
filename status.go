package main

import (
	"fmt"
)

type statusMap map[string]int

var status = make(statusMap)

const (
	ADDED    = "Added"
	REPLACED = "Replaced"
	DELETED  = "Deleted"
	CHECKED  = "Checked"
	SKIPPED  = "Skipped"
	ERROR    = "Error"
)

func (status statusMap) register(s string) {
	if v, ok := status[s]; ok {
		status[s] = v + 1
	} else {
		status[s] = 1
	}
}

func (status statusMap) report(s string, file string) {
	status.register(s)
	fmt.Println(string(s[0]), file)
}

func (status statusMap) reportError(e error) {
	status.register(ERROR)
	fmt.Println(e)
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
