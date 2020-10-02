package main

import "regexp"

type excludePatterns []*regexp.Regexp

func (excludes *excludePatterns) String() string {
	return ""
}

func (excludes *excludePatterns) Set(value string) error {
	*excludes = append(*excludes, regexp.MustCompile(value))
	return nil
}

func (excludes *excludePatterns) match(file string) bool {
	for _, patt := range *excludes {
		if patt.MatchString(file) {
			return true
		}
	}
	return false
}
