package main

import (
	"testing"
)

func TestExcludePatterns(t *testing.T) {
	var excludes excludePatterns

	excludes.Set("\\.ext$")
	excludes.Set("some")

	cases := map[string]bool{
		"some/file/name":        true, // matching 'some'
		"another/file/name.ext": true, // matching '.ext'
		"other/file/name":       false,
	}

	for file, result := range cases {
		if excludes.match(file) != result {
			if result {
				t.Errorf("Should have matched on %s\n", file)
			} else {
				t.Errorf("Unexpectedly matched on %s\n", file)
			}
		}
	}

}
