package dirsearch_test

import (
	"github.com/evilsocket/dirsearch"
	"math/rand"
	"testing"
	"time"
)

func init() {
	// seed RNG
	rand.Seed(time.Now().Unix())
}

func TestGetRandomUserAgent(t *testing.T) {
	results := make(map[string]bool)

	for i := 0; i < 75; i++ {
		ua := dirsearch.GetRandomUserAgent()
		results[ua] = true
	}

	//I'm a bit against testing randomness as the test result can float, but for such a small project I guess it's ok
	minimumExpected := 25

	if len(results) < minimumExpected {
		t.Errorf(
			"GetRandomUserAgent should have a good entropy level, got %d unique results, %d minimum expected",
			len(results),
			minimumExpected,
		)
	}
}
