package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

var (
	message            string
	nextURL            string
	forwardChance      float64
	errorChance        float64
	slowDownChance     float64
	bind               string
	googleCloudProject string
	limitedTrace       string
	enableTrace        string
)

func initParams() {
	message = os.Getenv("MESSAGE")
	nextURL = os.Getenv("NEXT_URL")
	bind = os.Getenv("BIND")
	googleCloudProject = os.Getenv("GOOGLE_CLOUD_PROJECT")
	limitedTrace = os.Getenv("LIMITED_TRACE")
	enableTrace = os.Getenv("ENABLE_TRACE")
	forwardChanceStr := os.Getenv("FORWARD_CHANCE")
	errorChanceStr := os.Getenv("ERROR_CHANCE")
	slowDownChanceStr := os.Getenv("SLOWDOWN_CHANCE")

	var err error

	if forwardChanceStr != "" {
		forwardChance, err = strconv.ParseFloat(forwardChanceStr, 64)
		if err != nil {
			log.Fatal("invalid FORWARD_CHANCE:", err)
		}
	}

	if errorChanceStr != "" {
		errorChance, err = strconv.ParseFloat(errorChanceStr, 64)
		if err != nil {
			log.Fatal("invalid ERROR_CHANCE:", err)
		}
	}

	if slowDownChanceStr != "" {
		slowDownChance, err = strconv.ParseFloat(slowDownChanceStr, 64)
		if err != nil {
			log.Fatal("invalid ERROR_CHANCE:", err)
		}
	}
}

func setupTrace() *http.Client {
	if enableTrace == "" {
		return http.DefaultClient
	}

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: googleCloudProject,
	})
	if err != nil {
		log.Fatal("create stackdriver trace exporter", err)
	}
	trace.RegisterExporter(exporter)

	if limitedTrace == "" {
		// will cause StackDriver spam under load!
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	return &http.Client{
		Transport: &ochttp.Transport{
			// Use Google Cloud propagation format.
			Propagation: &propagation.HTTPFormat{},
		},
	}
}

func main() {
	initParams()
	rand.Seed(time.Now().UnixNano())

	client := setupTrace()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if rand.Float64() < errorChance {
			http.Error(w, "unlucky draw, inject error", http.StatusTeapot)
			return
		}

		// not forwarding, just return the message
		if rand.Float64() > forwardChance || nextURL == "" {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "%s", message)
			return
		}

		req, err := http.NewRequest("GET", nextURL, nil)
		if err != nil {
			errStr := fmt.Sprintf("(forward: %s) failed to create request: %s", nextURL, err.Error())
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}

		// The trace ID from the incoming request will be
		// propagated to the outgoing request.
		req = req.WithContext(r.Context())

		resp, err := client.Do(req)
		if err != nil {
			errStr := fmt.Sprintf("(forward: %s) failed to get: %s", nextURL, err.Error())
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errStr := fmt.Sprintf("(forward: %s) bad status: %d", nextURL, resp.StatusCode)
			http.Error(w, errStr, resp.StatusCode)
			return
		}

		downStream, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errStr := fmt.Sprintf("(forward: %s) error reading body: %s", nextURL, err.Error())
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}

		if rand.Float64() < slowDownChance {
			time.Sleep(50 * time.Millisecond)
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s::%s", message, string(downStream))
		return
	})

	if bind == "" {
		bind = ":9090"
	}

	log.Fatal(http.ListenAndServe(bind, nil))
}
