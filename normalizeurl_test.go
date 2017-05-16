package dirsearch_test

import (
	"github.com/evilsocket/dirsearch"
	"testing"
)

var cases = map[string]string{
	"noschema.org/":            "http://noschema.org/",
	"http://nopath.org":        "http://nopath.org/",
	"http://imok.org/":         "http://imok.org/",
	"ihaveafile.com/index.php": "http://ihaveafile.com/index.php",
}

func TestNormalizeURL(t *testing.T) {
	for test, exp := range cases {
		dirsearch.NormalizeURL(&test)
		if test != exp {
			t.Errorf("Expected '%s', got '%s'.", exp, test)
		}
	}
}
