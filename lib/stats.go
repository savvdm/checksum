package lib

import (
	"fmt"
	"os"
)

type StatKey int

const (
	Visited = iota
	Added
	Replaced
	Deleted
	Checked
	Skipped
	Error
)

func (sk StatKey) String() string {
	return [...]string{"Visited", "Added", "Replaced", "Deleted", "Checked", "Skipped", "Error"}[sk]
}

type StatCounts [Error + 1]int

// increment the specified counter
func (stats *StatCounts) Register(sk StatKey) {
	stats[sk]++
}

// increment the specified counter and print trace with the file name
func (stats *StatCounts) Report(sk StatKey, file string) {
	stats.Register(sk)
	stats.ReportKey(sk, file)
}

// increment the specified counter and print trace if cond is true
func (stats *StatCounts) ReportIf(cond bool, sk StatKey, file string) {
	stats.Register(sk)
	if cond {
		stats.ReportKey(sk, file)
	}
}

// print trace with the file name
func (stats *StatCounts) ReportKey(sk StatKey, file string) {
	var label string
	if sk == Checked {
		label = "OK"
	} else {
		label := sk.String()
		label = string(label[0]) // use first (capital) letter as the label
	}
	fmt.Println(label, file)
}

// count and report error
func (stats *StatCounts) ReportError(e error) {
	stats.Register(Error)
	fmt.Fprintln(os.Stderr, e)
}

// print all stats
func (stats *StatCounts) Print() {
	for sk, count := range stats {
		fmt.Fprintf(os.Stderr, "%-12s%d\n", StatKey(sk).String()+":", count)
	}
}

func (stats *StatCounts) IsChanged() bool {
	return stats.sum([]StatKey{Added, Replaced, Deleted}) > 0
}

// calculate the sum of the specified stat counters
func (stats *StatCounts) sum(keys []StatKey) (count int) {
	for _, sk := range keys {
		count += stats[sk]
	}
	return
}
