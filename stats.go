package main

import (
	"fmt"
	"os"
)

type statKey int

const (
	Added = iota
	Replaced
	Deleted
	Ok
	Skipped
	Error
)

func (sk statKey) String() string {
	return [...]string{"Added", "Replaced", "Deleted", "OK", "Skipped", "Error"}[sk]
}

type statCounts [Error + 1]int

var stats statCounts

// increment the specified counter
func (stats *statCounts) register(sk statKey) {
	stats[sk]++
}

// increment the specified counter and print trace with the file name
func (stats *statCounts) report(sk statKey, file string) {
	stats.register(sk)
	stats.reportKey(sk, file)
}

// increment the specified counter and print trace if cond is true
func (stats *statCounts) reportIf(cond bool, sk statKey, file string) {
	stats.register(sk)
	if cond {
		stats.reportKey(sk, file)
	}
}

// print trace with the file name
func (stats *statCounts) reportKey(sk statKey, file string) {
	label := sk.String()
	if sk != Ok {
		label = string(label[0]) // use first (capital) letter as the label
	}
	fmt.Println(label, file)
}

// count and report error
func (stats *statCounts) reportError(e error) {
	stats.register(Error)
	fmt.Fprintln(os.Stderr, e)
}

// print all stats
func (stats *statCounts) print() {
	for sk, count := range stats {
		fmt.Fprintf(os.Stderr, "%-12s%d\n", statKey(sk).String()+":", count)
	}
}

// calculate the sum of the specified stat counters
func (stats *statCounts) sum(keys []statKey) (count int) {
	for _, sk := range keys {
		count += stats[sk]
	}
	return
}
