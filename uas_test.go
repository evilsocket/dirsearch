package main

import (
	"math/rand"
	"testing"
	"time"
)

func init() {
	// seed RNG
	rand.Seed(time.Now().Unix())
}

func TestGetRandomUserAgent(t *testing.T) {
	prev := ""
	for i := 0; i < 20; i++ {
		ua := GetRandomUserAgent()
		if ua == prev {
			t.Fatal("GetRandomUserAgent should never return the previous output.")
		}
		prev = ua
	}
}
