package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	app App
)

type App struct {
	addr       string
	liveDuring int
	started    time.Time
	worker     string
}

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
	hostname := fmt.Sprintf("Worker hostname: %s\n", app.worker)
	version := fmt.Sprintf("Version: %s\n", appVersion)
	gitCommitString := fmt.Sprintf("git Commit: %s\n", gitCommit)
	gitRepoString := fmt.Sprintf("git Repo: %s\n", gitRepo)
	buildDate := fmt.Sprintf("Build date: %s\n", buildStamp)

	requestDump = append(requestDump, hostname...)
	requestDump = append(requestDump, version...)
	requestDump = append(requestDump, gitCommitString...)
	requestDump = append(requestDump, gitRepoString...)
	requestDump = append(requestDump, buildDate...)

	log.Printf("remote_addr:%v, hostname: %v, method: %v, url:%v", r.RemoteAddr, r.Host, r.Method, r.URL)
	w.Write(requestDump)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	duration := time.Now().Sub(app.started)
	if duration.Seconds() > float64(app.liveDuring) {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}
}

func readiness(w http.ResponseWriter, r *http.Request) {
	duration := time.Now().Sub(app.started)
	if duration.Seconds() < 5 {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

func init() {
	flag.StringVar(&app.addr, "listen-address", ":8888",
		"The address to listen on for HTTP requests.")

	flag.IntVar(&app.liveDuring, "appTtl", 120,
		"How long app will be alive")
	app.started = time.Now()

	app.worker = func() string {
		host, _ := os.Hostname()
		return host
	}()
}

func main() {
	flag.Parse()

	prometheus.MustRegister(httpRequestsCount)

	// Create Server and Route Handlers
	r := mux.NewRouter()

	srv := &http.Server{
		Handler:      r,
		Addr:         app.addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	r.HandleFunc("/", formatRequest)
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/readiness", readiness)
	r.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting web server at %s\n", app.addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("http.ListenAndServer: %v\n", err)
		}
	}()

	waitForShutdown(srv)
}
