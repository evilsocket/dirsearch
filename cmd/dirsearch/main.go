// This software is a Go implementation of dirsearch by Mauro Soria
// (maurosoria at gmail dot com) written by Simone Margaritelli
// (evilsocket at gmail dot com).
package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/dirsearch"
	"github.com/fatih/color"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Stats will keep some statistics
// glued together.
type Stats struct {
	start    time.Time
	stop     time.Time
	total    time.Duration
	requests uint64
	errors   uint64
	oks      uint64
}

// Result represents the resul of each HEAD request
// to a given URL.
type Result struct {
	url      string
	status   int
	location string
	err      error
}

var (
	g = color.New(color.FgGreen)
	y = color.New(color.FgYellow)
	r = color.New(color.FgRed)
	b = color.New(color.FgBlue)

	stats = Stats{}

	base      = flag.String("url", "", "Base URL to start enumeration from.")
	ext       = flag.String("ext", "php", "File extension.")
	wordlist  = flag.String("wordlist", "dict.txt", "Wordlist file to use for enumeration.")
	consumers = flag.Uint("consumers", 8, "Number of concurrent consumers.")
	only200   = flag.Bool("200only", false, "If enabled, will only display responses with 200 status code.")
	maxerrors = flag.Uint64("maxerrors", 20, "Maximum number of errors to get before killing the program.")
)

func main() {
	setup()

	results := make(chan Result) // response consumer will listen on this channel
	urls := make(chan string)    // URLs will be pushed here
	wg := sync.WaitGroup{}       // Done condition

	fmt.Printf("Scanning %s with %d consumers ...\n\n", *base, *consumers)

	// start a fixed amount of consumers for URLs
	for i := uint(0); i < *consumers; i++ {
		go urlConsumer(urls, results)
	}

	// start the response consumer on a goroutine
	go respConsumer(results, &wg)

	stats.start = time.Now()

	// read wordlist line by line
	lines, err := dirsearch.LineReader(*wordlist)
	if err != nil {
		r.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	for line := range lines {
		wg.Add(1)
		// if present, replace '%EXT%' token with extension
		// and push the URL to the channel
		urls <- strings.Replace(*base+line, "%EXT%", *ext, -1)
	}

	// wait for everything to be completed
	wg.Wait()

	g.Println("\nDONE")

	printStats()
}

// Do some initialization.
// NOTE: We can't call this in the 'init' function otherwise
// flags are gonna be mandatory for unit test modules.
func setup() {
	flag.Parse()
	if *base == "" {
		flag.Usage()
		os.Exit(1)
	}

	// add schema
	if !strings.Contains(*base, "://") {
		*base = "http://" + *base
	}

	// add path
	if (*base)[len(*base)-1] != '/' {
		*base += "/"
	}

	// seed RNG
	rand.Seed(time.Now().Unix())

	// if interrupted, print statistics and exit
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		r.Println("\nINTERRUPTING ...")
		printStats()
		os.Exit(0)
	}()
}

// Print some stats
func printStats() {
	stats.stop = time.Now()
	stats.total = stats.stop.Sub(stats.start)

	fmt.Println("")
	fmt.Println("Requests :", stats.requests)
	fmt.Println("Errors   :", stats.errors)
	fmt.Println("200s     :", stats.oks)
	fmt.Println("Time     :", stats.total.Seconds(), "s")
	fmt.Println("Req/s    :", float64(stats.requests)/stats.total.Seconds())
}

// Consume URLs from the 'in' channel, execute HTTP HEAD request and pushes
// results to the out channel.
func urlConsumer(in <-chan string, out chan<- Result) {
	for url := range in {
		atomic.AddUint64(&stats.requests, 1)

		// Do not follow redirects.
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}}

		// Create HEAD request with random user agent.
		req, _ := http.NewRequest("HEAD", url, nil)
		req.Header.Set("User-Agent", dirsearch.GetRandomUserAgent())

		resp, err := client.Do(req)
		switch err {
		case nil:
			resp.Body.Close()
			if resp.StatusCode == 200 {
				atomic.AddUint64(&stats.oks, 1)
			}
			out <- Result{url, resp.StatusCode, resp.Header.Get("Location"), nil}
			break
		default:
			atomic.AddUint64(&stats.errors, 1)
			out <- Result{url, 500, "", err}
		}
	}
}

// Consume responses from a channel and print results.
func respConsumer(ch <-chan Result, wg *sync.WaitGroup) {
	for result := range ch {
		rps := fmt.Sprintf(" - ~%d r/s", int(float64(stats.requests)/time.Now().Sub(stats.start).Seconds()))
		now := time.Now().Format("15:04:05") + rps
		switch {
		// error not due to 404 response
		case result.err != nil && result.status != 404:
			r.Printf("[%s] [???] %s : %v\n", now, result.url, result.err)
		// 2xx
		case result.status >= 200 && result.status < 300:
			g.Printf("[%s] [%d] %s\n", now, result.status, result.url)
		// 3xx
		case !*only200 && result.status >= 300 && result.status < 400:
			b.Printf("[%s] [%d] %s -> %s\n", now, result.status, result.url, result.location)
		// 4xx
		case !*only200 && result.status >= 400 && result.status < 500 && result.status != 404:
			y.Printf("[%s] [%d] %s\n", now, result.status, result.url)
		// 5xx
		case !*only200 && result.status >= 500 && result.status < 600:
			r.Printf("[%s] [%d] %s\n", now, result.status, result.url)
		}

		if stats.errors > *maxerrors {
			r.Printf("\nExceeded %d errors, quitting ...\n", *maxerrors)
			os.Exit(1)
		}

		wg.Done()
	}
}
