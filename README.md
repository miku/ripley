README
======

Replay SOLR queries parsed from nginx logs.

Download Linux 64-bit binary: [ripley](https://github.com/miku/ripley/releases/download/v0.1.3/ripley).

Usage
-----

Input log format is the default nginx combined [format](https://github.com/miku/ripley/blob/8437e9bd241eb2605b0c6132095d4fdf84db0e82/cmd/ripley/main.go#L21):

    $ zcat log-20151124.gz
    112.101.12.10 - - [23/Nov/2015:11:01:50 +0100] "GET /solr/biblio/select?q=%28hi HTTP/1.1" 200 43 "-" "SLR"
    ...

We expect the `/solr/biblio/select` prefix.

Use `-addr` to point to a SOLR to warm up, `-run` flag to actually run the queries.

    $ zcat log-20151124.gz | ripley -addr 10.10.110.7:8085 -run
    {"elapsed":0.270542267,"status":"200 OK","url":"http://172..."}
    {"elapsed":0.287049684,"status":"200 OK","url":"http://172..."}
    ...

If `-ignore` is set, continue in the case of HTTP errors. Run queries in parallel with `-w` parameter:

    $ zcat log-20151124.gz | ripley -addr 10.10.110.7:8085 -ignore -w 8 -run
    {"elapsed":0.270542267,"status":"200 OK","url":"http://172..."}
    {"elapsed":0.287049684,"status":"200 OK","url":"http://172..."}
    ...

To get timing information out (in seconds), use [jq](https://stedolan.github.io/jq/):

    $ zcat log-20151124.gz | ripley -addr 10.10.110.7:8085 -run | jq -r '.elapsed'
    0.26466516700000003
    0.22494069300000002
    6.027045224
    6.151381966
    0.24651203600000002
    0.23020700700000002
    0.24007809400000002
    0.240172403
    0.24043798600000002
    0.240256606
    ...

If you need stats, you might want [st](https://github.com/nferraz/st) or something like it:

    $ zcat log-20151124.gz | ripley -ignore -w 4 -addr 10.10.110.7:8085 -run | jq -r '.elapsed' | st
    N       min         max     sum     mean    stddev  stderr
    2220    0.00110086  118.637 5548.31 2.49924 10.0882 0.214111
