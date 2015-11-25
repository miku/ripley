package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/satyrius/gonx"
)

const DefaultLogFormat = `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"`

type Opts struct {
	addr   string
	ignore bool
}

type Stats struct {
}

func worker(queue chan string, opts Opts, wg *sync.WaitGroup) {
	defer wg.Done()

	for link := range queue {

		start := time.Now()
		resp, err := http.Get(link)
		duration := time.Since(start)

		b, err := json.Marshal(map[string]interface{}{
			"url":     link,
			"status":  resp.Status,
			"elapsed": duration.Seconds(),
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))

		if resp.StatusCode >= 400 && !opts.ignore {
			log.Fatalf("%v", resp)
		}
		if err != nil && !opts.ignore {
			log.Fatal(err)
		}
		if _, err = ioutil.ReadAll(resp.Body); err != nil && !opts.ignore {
			log.Fatal(err)
		}
		resp.Body.Close()
	}
}

func main() {

	addr := flag.String("addr", "", "hostport of SOLR to query")
	run := flag.Bool("run", false, "actually run the requests")
	ignore := flag.Bool("ignore", false, "ignore errors")
	w := flag.Int("w", 1, "number of requests to run in parallel")

	flag.Parse()

	reader := gonx.NewReader(os.Stdin, DefaultLogFormat)

	queue := make(chan string)
	var wg sync.WaitGroup

	opts := Opts{addr: *addr, ignore: *ignore}

	for i := 0; i < *w; i++ {
		wg.Add(1)
		go worker(queue, opts, &wg)
	}

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		f, err := rec.Field("request")
		if err != nil {
			log.Fatal(err)
		}

		// skip non GET requests
		if !strings.HasPrefix(f, "GET") {
			continue
		}

		// extract path
		replacer := strings.NewReplacer("GET ", "", "HTTP/1.1", "", "HTTP/1.0", "")
		p := replacer.Replace(f)

		// only consider solr requests
		if !strings.HasPrefix(p, "/solr/biblio/select") {
			continue
		}

		link := path.Join(*addr, p)

		if !strings.HasPrefix(link, "http") && *addr != "" {
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
}
