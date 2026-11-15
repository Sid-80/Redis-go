package main

import (
	"fmt"
	"log"

	"github.com/Sid-80/Redis-go/store"
	"github.com/Sid-80/Redis-go/store/kv"
)

func main() {
	var base store.Storer = kv.NewStore(20)

	// Middlewares wrap each other because they all implement store.Storer:
	// calls go through metrics first, then reach the kv store.
	// To see every call logged, wrap it: s := NewLoggingMiddleware(metrics)
	metrics := NewMetricMiddleware(base)
	var s store.Storer = metrics

	if err := s.Set("pageViews", "0"); err != nil {
		log.Fatalf("initial SET failed: %v", err)
	}

	const hits = 1000
	cmds := make([]Command, 0, hits+2)
	for range hits {
		cmds = append(cmds, Command{Op: "INCR", Key: "pageViews"})
	}
	// Two bad commands to show error handling.
	cmds = append(cmds, Command{Op: "INCR", Key: ""}, Command{Op: "WAT", Key: "pageViews"})

	for _, err := range restartsOnBoot(s, cmds) {
		fmt.Println("Error:", err)
	}

	got, err := s.Get("pageViews")
	if err != nil {
		log.Fatalf("GET failed: %v", err)
	}

	fmt.Println("Got: ", got)
	fmt.Println("Want:", hits)
	fmt.Println()
	metrics.Report()
}
