package main

import (
	"fmt"
)

type statKey int

const (
	Added = iota
	Replaced
	Deleted
	Checked
	Skipped
	Error
)

func (sk statKey) String() string {
	return [...]string{"Added", "Replaced", "Deleted", "Checked", "Skipped", "Error"}[sk]
}

type statCounts [Error + 1]int

var stats statCounts

func (stats *statCounts) register(sk statKey) {
	stats[sk]++
}

func (stats *statCounts) report(sk statKey, file string) {
	stats.register(sk)
	label := sk.String()
	fmt.Println(string(label[0]), file)
}

func (stats *statCounts) reportError(e error) {
	stats.register(Error)
	fmt.Println(e)
}

func (stats *statCounts) print() {
	for sk, count := range stats {
		fmt.Printf("%s: %d\n", statKey(sk).String(), count)
	}
}

func (stats *statCounts) sum(keys []statKey) (count int) {
	for _, sk := range keys {
		count += stats[sk]
	}
	return
}
