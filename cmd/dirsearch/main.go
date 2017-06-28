// This software is a Go implementation of dirsearch by Mauro Soria
// (maurosoria at gmail dot com) written by Simone Margaritelli
// (evilsocket at gmail dot com).
package main

import (
	"flag"
	"fmt"
	"github.com/evilsocket/brutemachine"
	"github.com/evilsocket/dirsearch"
	"github.com/fatih/color"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"strings"
	"syscall"
	"time"
)

// Result represents the resul of each HEAD request
// to a given URL.
type Result struct {
	url      string
	status   int
	location string
	err      error
}

var (

	m *brutemachine.Machine

	errors = uint64(0)

	g = color.New(color.FgGreen)
	y = color.New(color.FgYellow)
	r = color.New(color.FgRed)
	b = color.New(color.FgBlue)

	base      = flag.String("url", "", "Base URL to start enumeration from.")
	ext       = flag.String("ext", "php", "File extension.")
	wordlist  = flag.String("wordlist", "dict.txt", "Wordlist file to use for enumeration.")
	consumers = flag.Int("consumers", 8, "Number of concurrent consumers.")
	only200   = flag.Bool("200only", false, "If enabled, will only display responses with 200 status code.")
	maxerrors = flag.Uint64("maxerrors", 20, "Maximum number of errors to get before killing the program.")
)

func DoRequest(page string) interface{} {
	url := strings.Replace(fmt.Sprintf("%s%s", *base, page), "%EXT%", *ext, -1)
	// Do not follow redirects.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	// Create HEAD request with random user agent.
	req, _ := http.NewRequest("HEAD", url, nil)
	req.Header.Set("User-Agent", dirsearch.GetRandomUserAgent())

	if resp, err := client.Do(req); err == nil {
		resp.Body.Close()
		if resp.StatusCode == 200 || !*only200 {
			return Result{url, resp.StatusCode, resp.Header.Get("Location"), nil}
		}
	} else {
		atomic.AddUint64(&errors, 1)
	}

	return nil
}

func OnResult(res interface{}) {
	result, ok := res.(Result)
	if !ok {
		r.Printf( "Error while converting result.\n" )
		return
	}
	
	now := time.Now().Format("15:04:05") 	
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
	case result.status >= 400 && result.status < 500 && result.status != 404:
		y.Printf("[%s] [%d] %s\n", now, result.status, result.url)
	// 5xx
	case result.status >= 500 && result.status < 600:
		r.Printf("[%s] [%d] %s\n", now, result.status, result.url)
	}

	if errors > *maxerrors {
		r.Printf("\nExceeded %d errors, quitting ...\n", *maxerrors)
		os.Exit(1)
	}
}

func main() {
	setup()

	m = brutemachine.New( *consumers, *wordlist, DoRequest, OnResult)
    if err := m.Start(); err != nil {
        panic(err)
    }

    m.Wait()

	g.Println("\nDONE")

	printStats()
}

// Do some initialization.
// NOTE: We can't call this in the 'init' function otherwise
// flags are gonna be mandatory for unit test modules.
func setup() {
	flag.Parse()

	if err := dirsearch.NormalizeURL(base); err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
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
	m.UpdateStats()

	fmt.Println("")
	fmt.Println("Requests :", m.Stats.Execs)
	fmt.Println("Errors   :", errors)
	fmt.Println("Results  :", m.Stats.Results)
	fmt.Println("Time     :", m.Stats.Total.Seconds(), "s")
	fmt.Println("Req/s    :", m.Stats.Eps)
}

