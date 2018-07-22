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
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type Result struct {
	url      string
	status   int
	size     uint64
	location string
	err      error
}

var (
	m *brutemachine.Machine

	g = color.New(color.FgGreen)
	y = color.New(color.FgYellow)
	r = color.New(color.FgRed)
	b = color.New(color.FgBlue)

	errors     = uint64(0)
	fail_codes = make(map[int]bool)

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient = &http.Client{
		Transport: transport,
	}

	exclude   = flag.String("x", "404", "Status codes to exclude")
	method    = flag.String("m", "GET", "Request method (HEAD / GET)")
	base      = flag.String("u", "", "URL to enumerate")
	wordlist  = flag.String("w", "dict.txt", "Wordlist file")
	maxerrors = flag.Uint64("E", 10, "Max. errors before exiting")
	timeout   = flag.Uint("T", 10, "Timeout before killing the request")
	wildcard  = flag.Bool("sw", false, "Skip wildcard responses")
	only200   = flag.Bool("2", false, "If enabled, will only display responses with 200 status code.")
	follow    = flag.Bool("f", false, "Follow redirects.")
	threads   = flag.Int("t", 10, "Number of concurrent threads.")
)

func IsWildcard(url string) bool {
	test := uuid.Must(uuid.NewV4(), nil).String()
	res, ok := DoRequest(test).(Result)

	if !ok {
		return false
	}

	if res.status == 200 || !fail_codes[res.status] {
		*wildcard = false
		return true
	}

	*wildcard = false
	return false
}

func DoRequest(page string) interface{} {
	url := fmt.Sprintf("%s%s", *base, page)

	req, _ := http.NewRequest(*method, url, nil)

	req.Header.Set("User-Agent", dirsearch.GetRandomUserAgent())
	req.Header.Set("Accept", "*/*")

	resp, err := httpClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		atomic.AddUint64(&errors, 1)
		return Result{url, 0, uint64(0), "", err}
	}

	content, _ := ioutil.ReadAll(resp.Body)

	if (resp.StatusCode == 200 && *only200) || (!fail_codes[resp.StatusCode] && !*only200) || (*wildcard) {
		return Result{url, resp.StatusCode, uint64(len(content)), resp.Header.Get("Location"), nil}
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
		r.Fprintf(os.Stderr, "[%s] %s : %v\n", now, result.url, result.err)

	// 2xx
	case result.status >= 200 && result.status < 300:
		if *method == "GET" {
			g.Printf("[%s] %-3d %-9d %s\n", now, result.status, result.size, result.url)
		} else {
			g.Printf("[%s] %-3d %s\n", now, result.status, result.url)
		}

	// 3xx
	case !*only200 && result.status >= 300 && result.status < 400:
		b.Printf("[%s] %-3d %-9d %s -> %s\n", now, result.status, result.size, result.url, result.location)

	// 4xx
	case result.status >= 400 && result.status < 500 && result.status != 404:
		y.Printf("[%s] %-3d %-9d %s\n", now, result.status, result.size, result.url)

	// 5xx
	case result.status >= 500 && result.status < 600:
		r.Printf("[%s] %-3d %-9d %s\n", now, result.status, result.size, result.url)
	}

	if errors > *maxerrors {
		r.Fprintf(os.Stderr, "\nExceeded %d errors, quitting ...", *maxerrors)
		os.Exit(1)
	}
}

func main() {
	setup()

	// create a list of exclusions
	if *exclude != "" {
		for _, x := range strings.Split(*exclude, ",") {
			y, _ := strconv.Atoi(x)
			fail_codes[y] = true
		}
	}

	// set timeout
	httpClient.Timeout = time.Duration(*timeout) * time.Second

	// set redirects policy
	if !*follow {
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// check for wildcard responses, and return if true
	if *wildcard == true {
		if IsWildcard(*base) == true {
			r.Fprintf(os.Stderr, "\nWildcard detected on %s, skipping....\n", *base)
			os.Exit(0)
		}
	}

	// start
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
// *are gonna be mandatory for unit test modules.
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
