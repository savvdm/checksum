package lib

import (
	"fmt"
	"os"
)

type Status byte

const (
	Read = iota
	Visited
	Added
	Replaced
	Deleted
	Checked
	Skipped
	Error
)

func (status Status) String() string {
	return [...]string{"Read", "Visited", "Added", "Replaced", "Deleted", "Checked", "Skipped", "Error"}[status]
}

// print trace with the file name
func ReportFile(file string, status Status) {
	if status > Visited { // NOTE: Read & Visited are not reported
		var label string
		if status == Checked {
			label = "OK"
		} else {
			label = status.String()
			label = string(label[0]) // use first (capital) letter as the label
		}
		fmt.Println(label, file)
	}
}

func ReportStatus(file string, status Status, verbose bool) {
	switch status {
	case Added, Replaced, Deleted:
		ReportFile(file, status)
	case Checked:
		if verbose {
			ReportFile(file, status)
		}
	}
}

type StatCounts [Error + 1]int

// increment the specified counter
func (stats *StatCounts) Register(status Status) {
	stats[status]++
}

// directly set the specified counter
// NOTE: override any previous value
func (stats *StatCounts) Set(status Status, count int) {
	stats[status] = count
}

// increment the specified counter and print trace with the file name
func (stats *StatCounts) Report(status Status, file string) {
	stats.Register(status)
	ReportFile(file, status)
}

// increment the specified counter and print trace if cond is true
func (stats *StatCounts) ReportIf(cond bool, status Status, file string) {
	stats.Register(status)
	if cond {
		ReportFile(file, status)
	}
}

// count and report error
func (stats *StatCounts) ReportError(e error) {
	stats.Register(Error)
	fmt.Fprintln(os.Stderr, e)
}

// print all stats
func (stats *StatCounts) Print() {
	for status, count := range stats {
		switch status {
		case Read, Visited:
			// don't print those
		default:
			fmt.Fprintf(os.Stderr, "%-12s%d\n", Status(status).String()+":", count)
		}
	}
}

func (stats *StatCounts) HasChanged() bool {
	return stats.sum([]Status{Added, Replaced, Deleted}) > 0
}

// calculate the sum of the specified stat counters
func (stats *StatCounts) sum(keys []Status) (count int) {
	for _, status := range keys {
		count += stats[status]
	}
	return
}
