package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	u "github.com/agolebiowska/distributed-counter/utils"
)

func main() {
	l := log.New(os.Stdout, "counter", log.LstdFlags)

	me, err := os.Hostname()
	if err != nil {
		l.Fatal("[ERROR] Cannot obtain hostname:", err.Error())
	}

	data := []byte(me)
	items := Items{}
	err = u.Do(http.MethodPost, "http://coordinator/counters", items, bytes.NewBuffer(data))
	if err != nil {
		l.Fatal("[ERROR] Cannot add counter: ", err.Error())
	}

	c := NewCounter(me, items)

	sm := http.NewServeMux()
	sm.Handle("/count/", NewCountItems(l, c))
	sm.Handle("/init", NewInit(l, c))
	sm.Handle("/abort", NewAbort(l, c))
	sm.Handle("/commit", NewCommit(l, c))

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
			log.Fatal(err)
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
