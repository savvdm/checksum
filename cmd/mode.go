package cmd

import "errors"

type CheckMode int

const (
	New = iota
	Modified
	All
)

func (mode CheckMode) String() string {
	return [...]string{"new", "modified", "all"}[mode]
}

func (mode *CheckMode) Set(value string) error {
	switch {
	case value == CheckMode(New).String():
		*mode = New
	case value == CheckMode(Modified).String():
		*mode = Modified
	case value == CheckMode(All).String():
		*mode = All
	default:
		return errors.New("")
	}
	return nil
}
