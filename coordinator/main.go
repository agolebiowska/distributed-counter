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

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	go func() {
		for {
			for range time.Tick(10 * time.Second) {
				for _, counter := range c.Counters {
					url := fmt.Sprintf("http://%s/health", counter.Addr)
					resp, err := c.Do(http.MethodGet, url, nil)
					if err != nil || resp.StatusCode != 200 {
						if c.Counters[counter.Addr].RecoveryTries >= 5 {
							c.removeCounter(counter.Addr)
							l.Printf("[INFO] %s removed", counter.Addr)
							continue
						}
						l.Printf("[INFO] %s not responding", counter.Addr)
						c.Counters[counter.Addr].IsDead = true
						c.Counters[counter.Addr].RecoveryTries++
					}
				}
			}
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
