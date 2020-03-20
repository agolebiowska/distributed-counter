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
	// @todo: "apply" to coordinator with our address which gets us all items?
	me, err := os.Hostname()
	if err != nil {
		log.Fatal("[ERROR] Cannot obtain hostname:", err.Error())
	}

	data := []byte(me)
	items := Items{}
	err = u.Do(http.MethodPost, "http://localhost:8080/counters", items, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("[ERROR] Cannot add counter:", err.Error())
	}

	c := NewCounter(items)
	log.Println(c)
	log.Println(me)

	sm := http.NewServeMux()

	s := &http.Server{
		Addr:         ":8080",
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
	log.Println("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
