/*
 * This software is a Go implementation of dirsearch by Mauro Soria
 * (maurosoria at gmail dot com) written by Simone Margaritelli
 * (evilsocket at gmail dot com).
 */
package main

import (
	"bufio"
	"flag"
	"fmt"
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

// This structure will keep some statistics
// glued together.
type Stats struct {
	start    time.Time
	stop     time.Time
	total    time.Duration
	requests uint64
	errors   uint64
	oks      uint64
}

// Represents the resul of each HEAD request
// to a given URL.
type Result struct {
	url      string
	status   int
	location string
	err      error
}

var (
	// for each request, a random UA will be selected from this list
	uas = [...]string{
		"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.1 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2226.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1",
		"Mozilla/5.0 (Windows NT 6.3; rv:36.0) Gecko/20100101 Firefox/36.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10; rv:33.0) Gecko/20100101 Firefox/33.0",
		"Mozilla/5.0 (X11; Linux i586; rv:31.0) Gecko/20100101 Firefox/31.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:31.0) Gecko/20130401 Firefox/31.0",
		"Mozilla/5.0 (Windows NT 5.1; rv:31.0) Gecko/20100101 Firefox/31.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko",
		"Mozilla/5.0 (compatible, MSIE 11, Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (compatible; MSIE 10.6; Windows NT 6.1; Trident/5.0; InfoPath.2; SLCC1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 2.0.50727) 3gpp-gba UNTRUSTED/1.0",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 7.0; InfoPath.3; .NET CLR 3.1.40767; Trident/6.0; en-IN)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/6.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/5.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/4.0; InfoPath.2; SV1; .NET CLR 2.0.50727; WOW64)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Macintosh; Intel Mac OS X 10_7_3; Trident/6.0)",
		"Mozilla/4.0 (Compatible; MSIE 8.0; Windows NT 5.2; Trident/6.0)",
		"Mozilla/4.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/5.0)",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; WIndows NT 9.0; en-US))",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 7.1; Trident/5.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; Media Center PC 6.0; InfoPath.3; MS-RTC LM 8; Zune 4.7)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; Media Center PC 6.0; InfoPath.3; MS-RTC LM 8; Zune 4.7",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; Zune 4.0; InfoPath.3; MS-RTC LM 8; .NET4.0C; .NET4.0E)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; chromeframe/12.0.742.112)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 2.0.50727; SLCC2; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; Zune 4.0; Tablet PC 2.0; InfoPath.3; .NET4.0C; .NET4.0E)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; yie8)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0; GTB7.4; InfoPath.2; SV1; .NET CLR 3.3.69573; WOW64; en-US)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; InfoPath.1; SV1; .NET CLR 3.8.36217; WOW64; en-US)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; .NET CLR 2.7.58687; SLCC2; Media Center PC 5.0; Zune 3.4; Tablet PC 3.6; InfoPath.3)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.2; Trident/4.0; Media Center PC 4.0; SLCC1; .NET CLR 3.0.04320)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; SLCC1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; InfoPath.2; SLCC1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; SLCC1; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.0; Trident/4.0; InfoPath.1; SV1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 5.0; Trident/4.0; FBSMTWB; .NET CLR 2.0.34861; .NET CLR 3.0.3746.3218; .NET CLR 3.5.33652; msn OptimizedIE8;ENUS)",
		"Mozilla/4.0(compatible; MSIE 7.0b; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; Media Center PC 3.0; .NET CLR 1.0.3705; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.1)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; FDM; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; en-US)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; el-GR)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 5.2)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP)",
		"Mozilla/4.0 (compatible; MSIE 6.01; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.1; DigExt)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.1)",
	}

	nuas = len(uas)

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

func init() {
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

func main() {
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
	err, lines := lineReader(*wordlist)
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
		req.Header.Set("User-Agent", uas[rand.Intn(nuas)])

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
		defer wg.Done()

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
	}
}

// lineReader will accept the name of a file as argument
// and will return a channel from which lines can be read
// one at a time.
func lineReader(filename string) (error, chan string) {
	out := make(chan string)
	fp, err := os.Open(filename)
	if err != nil {
		return err, nil
	}

	go func() {
		defer fp.Close()
		// we need to close the out channel in order
		// to signal the end-of-data condition
		defer close(out)
		scanner := bufio.NewScanner(fp)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	return nil, out
}
