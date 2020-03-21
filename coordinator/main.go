package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var l = log.New(os.Stdout, "coordinator-", log.LstdFlags)

func main() {
	c := NewCoordinator()

	sm := http.NewServeMux()
	sm.Handle("/items/", NewItemsCount(c))
	sm.Handle("/items", NewItemsAdd(c))
	sm.Handle("/counters", NewCounterAdd(c))
	sm.Handle("/health", NewHealthCheck())

	s := &http.Server{
		Addr:         ":80",
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go check(c.Counters)

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan
	l.Println("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}

func check(c []*Counter) {
	for {
		for range time.Tick(30 * time.Second) {
			for i, counter := range c {
				url := fmt.Sprintf("http://%s/health", counter.Addr)
				resp, err := Do(http.MethodGet, url, nil)
				if err != nil {
					l.Printf("[INFO] Health check failed: %s", err.Error())

				}
				if resp.StatusCode != 200 {
					c = append(c[:i], c[i+1:]...)
					l.Printf("[INFO] Removed %s from counters", counter.Addr)
				}
			}
		}
	}
}
