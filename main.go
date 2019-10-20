package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

var (
	addr = flag.String("listen-address", ":8888",
		"The address to listen on for HTTP requests.")
	liveDuring = flag.Int("appTtl", 120,
		"How long app will be alive")
	started = time.Now()
)

var (
	httpRequestsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Requests count per location",
		}, []string{"location"})
)

func formatRequest(w http.ResponseWriter, r *http.Request) {
	// Save a copy of this request for debugging.
	httpRequestsCount.WithLabelValues("/").Inc()
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("remote_addr:%v, hostname: %v, method: %v, url:%v", r.RemoteAddr, r.Host, r.Method, r.URL)
	w.Write(requestDump)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	duration := time.Now().Sub(started)
	if duration.Seconds() > float64(*liveDuring) {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}
}

func readiness(w http.ResponseWriter, r *http.Request) {
	duration := time.Now().Sub(started)
	if duration.Seconds() > 5 {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}
}

func main() {
	flag.Parse()

	prometheus.MustRegister(httpRequestsCount)

	http.HandleFunc("/", formatRequest)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/readiness", readiness)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting web server at %s\n", *addr)

	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Printf("http.ListenAndServer: %v\n", err)
	}
}
