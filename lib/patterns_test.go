package lib

import (
	"testing"
)

func TestPatterns(t *testing.T) {
	var patts Patterns

	patts.Set("\\.ext$")
	patts.Set("some")

	cases := map[string]bool{
		"some/file/name":        true, // matching 'some'
		"another/file/name.ext": true, // matching '.ext'
		"other/file/name":       false,
	}

	for file, result := range cases {
		if patts.Match(file) != result {
			if result {
				t.Errorf("Should have matched on %s\n", file)
			} else {
				t.Errorf("Unexpectedly matched on %s\n", file)
			}
		}
	}

}
