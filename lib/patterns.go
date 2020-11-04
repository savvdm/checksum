package lib

import "regexp"

type Patterns []*regexp.Regexp

func (patts *Patterns) String() string {
	return ""
}

func (patts *Patterns) Set(value string) error {
	*patts = append(*patts, regexp.MustCompile(value))
	return nil
}

func (patts *Patterns) Match(file string) bool {
	for _, patt := range *patts {
		if patt.MatchString(file) {
			return true
		}
	}
	return false
}
