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
				for i, counter := range c.Counters {
					url := fmt.Sprintf("http://%s/health", counter.Addr)
					resp, err := c.Do(http.MethodGet, url, nil)
					if err != nil || resp.StatusCode != 200 {
						if c.Counters[i].RecoveryTries >= 5 {
							c.Counters = append(c.Counters[:i], c.Counters[i+1:]...)
							l.Printf("[INFO] %s removed", counter.Addr)
							continue
						}
						c.Counters[i].IsDead = true
						c.Counters[i].RecoveryTries++
						l.Printf("[INFO] %s not query able", counter.Addr)
						continue
					}
					if c.Counters[i].IsDead {
						c.Counters[i].IsDead = false
						c.Counters[i].RecoveryTries = 0
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
