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

func main() {
	l := log.New(os.Stdout, "coordinator-", log.LstdFlags)
	c := NewCoordinator()

	sm := http.NewServeMux()
	sm.Handle("/items/", NewItemsCount(l, c))
	sm.Handle("/items", NewItemsAdd(l, c))
	sm.Handle("/counters", NewCounterAdd(l, c))
	sm.Handle("/health", NewHealthCheck(l))

	s := &http.Server{
		Addr:         ":80",
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		for {
			for range time.Tick(30 * time.Second) {
				for i, counter := range c.Counters {
					url := fmt.Sprintf("http://%s/health", counter.Addr)
					resp, err := Do(http.MethodGet, url, nil)
					if err != nil {
						l.Printf("[INFO] Health check failed: %s", err.Error())

					}
					if resp.StatusCode != 200 {
						c.Counters = append(c.Counters[:i], c.Counters[i+1:]...)
						l.Printf("[INFO] Removed %s from counters", counter.Addr)
					}
				}
			}
		}
	}()

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
