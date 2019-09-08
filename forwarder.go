package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

var (
	message       string
	nextURL       string
	forwardChance float64
	bind          string
)

func initParams() {
	message = os.Getenv("MESSAGE")
	nextURL = os.Getenv("NEXT_URL")
	bind = os.Getenv("BIND")
	forwardChanceStr := os.Getenv("FORWARD_CHANCE")
	if forwardChanceStr == "" {
		return
	}

	var err error
	forwardChance, err = strconv.ParseFloat(forwardChanceStr, 64)
	if err != nil {
		log.Fatal("invalid FORWARD_CHANCE:", err)
	}
}

func main() {
	initParams()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// not forwarding, just return the message
		if rand.Float64() > forwardChance || nextURL == "" {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "%s", message)
			return
		}

		resp, err := http.Get(nextURL)
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
			http.Error(w, errStr, resp.StatusCode)
			return
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
