// This software is a Go implementation of dirsearch by Mauro Soria
// (maurosoria at gmail dot com) written by Simone Margaritelli
// (evilsocket at gmail dot com).

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/eur0pa/dirsearch-go"
	"github.com/evilsocket/brutemachine"
	"github.com/fatih/color"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// Result represents the resul of each HEAD request
// to a given URL.
type Result struct {
	url      string
	status   int
	size     int
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
	threads   = flag.Int("threads", 8, "Number of concurrent threads.")
	only200   = flag.Bool("200only", false, "If enabled, will only display responses with 200 status code.")
	maxerrors = flag.Uint64("maxerrors", 20, "Maximum number of errors to get before killing the program.")
	timeout   = flag.Uint("timeout", 5, "Timeout before killing the request.")
	wildcard  = flag.Bool("wild", false, "Test and skip wildcard responses.")
	method    = flag.String("method", "GET", "Request method (HEAD / GET)")
)

func IsWildcard(url string) bool {
	test := uuid.Must(uuid.NewV4(), nil).String()
	res, ok := DoRequest(test).(Result)

	if !ok {
		return false
	}

	if res.status == 200 {
		*wildcard = false
		return true
	}

	*wildcard = false
	return false
}

func DoRequest(page string) interface{} {
	/*if errors > *maxerrors {
		r.Fprintf(os.Stderr, "\nExceeded %d errors, quitting ...", *maxerrors)
		os.Exit(1)
	}*/

	url := strings.Replace(fmt.Sprintf("%s%s", *base, page), "%EXT%", *ext, -1)

	// Do not verify certificates, do not follow redirects.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(*timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

	// Create request with random user agent.
	req, _ := http.NewRequest(*method, url, nil)
	req.Header.Set("User-Agent", dirsearch.GetRandomUserAgent())
	req.Header.Set("Accept", "*/*")

	if resp, err := client.Do(req); err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 || !*only200 || *wildcard == true {
			content, _ := ioutil.ReadAll(resp.Body)
			return Result{url, resp.StatusCode, len(content), resp.Header.Get("Location"), nil}
		}
	} else {
		atomic.AddUint64(&errors, 1)
	}

	return nil
}

func OnResult(res interface{}) {
	result, ok := res.(Result)

	if !ok {
		r.Fprintln(os.Stderr, "Error while converting result.")
		return
	}

	now := time.Now().Format("15:04:05")

	switch {
	// error not due to 404 response
	case result.err != nil && result.status != 404:
		r.Printf("[%s] ........ %s : %v\n", now, result.url, result.err)
	// 2xx
	case result.status >= 200 && result.status < 300:
		if *method == "GET" {
			g.Printf("[%s] %-3d %-8d %s\n", now, result.status, result.size, result.url)
		} else {
			g.Printf("[%s] %-3d %s\n", now, result.status, result.url)
		}
	// 3xx
	case !*only200 && result.status >= 300 && result.status < 400:
		b.Printf("[%s] %-3d %s -> %s\n", now, result.status, result.url, result.location)
	// 4xx
	case result.status >= 400 && result.status < 500 && result.status != 404:
		y.Printf("[%s] %-3d %s\n", now, result.status, result.url)
	// 5xx
	case result.status >= 500 && result.status < 600:
		r.Printf("[%s] %-3d %s\n", now, result.status, result.url)
	}

}

func main() {
	setup()

	// check for wildcard responses, and return if true
	if *wildcard == true {
		if IsWildcard(*base) == true {
			r.Fprintf(os.Stderr, "\nWildcard detected on %s, skipping....\n", *base)
			os.Exit(0)
		}
	}

	m = brutemachine.New(*threads, *wordlist, DoRequest, OnResult)

	if err := m.Start(); err != nil {
		panic(err)
	}

	m.Wait()

	g.Fprintln(os.Stderr, "\nDONE")

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
		r.Fprintln(os.Stderr, "\nINTERRUPTING ...")
		printStats()
		os.Exit(0)
	}()
}

// Print some stats
func printStats() {
	m.UpdateStats()

	fmt.Fprintln(os.Stderr, "Requests :", m.Stats.Execs)
	fmt.Fprintln(os.Stderr, "Errors   :", errors)
	fmt.Fprintln(os.Stderr, "Results  :", m.Stats.Results)
	fmt.Fprintln(os.Stderr, "Time     :", m.Stats.Total.Seconds(), "s")
	fmt.Fprintln(os.Stderr, "Req/s    :", m.Stats.Eps)
}
