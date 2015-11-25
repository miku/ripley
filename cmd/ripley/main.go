package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/satyrius/gonx"
)

const (
	Version   = "0.1.1"
	LogFormat = `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"`
)

var replacer = strings.NewReplacer("GET ", "", "HTTP/1.1", "", "HTTP/1.0", "")

type Opts struct {
	addr   string
	ignore bool
}

func worker(queue chan string, out chan string, opts Opts, wg *sync.WaitGroup) {
	defer wg.Done()
	for link := range queue {
		start := time.Now()
		resp, err := http.Get(link)
		duration := time.Since(start)

		if err != nil {
			if opts.ignore {
				log.Println(err)
			} else {
				log.Fatal(err)
			}
		}

		b, err := json.Marshal(map[string]interface{}{
			"url":     link,
			"status":  resp.Status,
			"elapsed": duration.Seconds(),
		})
		if err != nil {
			log.Fatal(err)
		}
		out <- string(b)

		if resp.StatusCode >= 400 && !opts.ignore {
			log.Fatalf("%v", resp)
		}
		resp.Body.Close()
	}
}

func writer(out chan string, done chan bool) {
	for s := range out {
		fmt.Println(s)
	}
	done <- true
}

func main() {

	addr := flag.String("addr", "", "hostport of SOLR to query, e.g. 10.1.1.7:8085 or https://10.1.1.8:1234")
	run := flag.Bool("run", false, "actually run the requests")
	ignore := flag.Bool("ignore", false, "ignore errors")
	w := flag.Int("w", 1, "number of requests to run in parallel")
	version := flag.Bool("v", false, "show version and exit")

	flag.Parse()

	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	reader := gonx.NewReader(os.Stdin, LogFormat)

	queue := make(chan string)
	out := make(chan string)
	done := make(chan bool)
	go writer(out, done)

	opts := Opts{addr: *addr, ignore: *ignore}
	var wg sync.WaitGroup

	for i := 0; i < *w; i++ {
		wg.Add(1)
		go worker(queue, out, opts, &wg)
	}

	for {
		rec, err := reader.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// get the request field
		f, err := rec.Field("request")
		if err != nil {
			log.Fatal(err)
		}

		if !strings.HasPrefix(f, "GET") {
			continue
		}

		// we only want the path
		p := replacer.Replace(f)

		if !strings.HasPrefix(p, "/solr/biblio/select") {
			continue
		}

		link := path.Join(*addr, p)

		if !strings.HasPrefix(link, "http") {
			link = "http://" + link
		}

		if *run {
			queue <- link
		} else {
			log.Println(link)
		}
	}

	close(queue)
	wg.Wait()
	close(out)
	<-done
}
