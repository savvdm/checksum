package main

import "errors"

type checkMode int

const (
	New = iota
	Modified
	All
)

func (mode checkMode) String() string {
	return [...]string{"new", "modified", "all"}[mode]
}

func (mode *checkMode) Set(value string) error {
	switch {
	case value == checkMode(New).String():
		*mode = New
	case value == checkMode(Modified).String():
		*mode = Modified
	case value == checkMode(All).String():
		*mode = All
	default:
		return errors.New("")
	}
	return nil
}
