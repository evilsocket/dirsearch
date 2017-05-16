package dirsearch_test

import (
	"github.com/evilsocket/dirsearch"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const (
	tokens        = "foo\nbar\nboo\nmuu\neof\n"
	test_filename = "/tmp/dirsearchtest"
)

func setup() {
	// create file used for testing
	if err := ioutil.WriteFile(test_filename, []byte(tokens), 0644); err != nil {
		panic(err)
	}
}

func teardown() {
	if err := os.Remove(test_filename); err != nil {
		panic(err)
	}
}

func TestLineReader(t *testing.T) {
	setup()
	defer teardown()

	lines, err := dirsearch.LineReader(test_filename, 0)
	if err != nil {
		t.Errorf("LineReader should not fail with error: %v", err)
	}

	expected := strings.Split(tokens, "\n")
	nexpected := len(expected)
	i := 0

	for line := range lines {
		if i >= nexpected {
			t.Errorf("Test file and array should contain the same number of strings.")
		} else if expected[i] != line {
			t.Errorf("Got '%s', expected '%s'.", line, expected[i])
		}
		i++
	}
}
