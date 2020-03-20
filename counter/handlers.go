package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Init struct {
	log     *log.Logger
	counter *Counter
}

type Abort struct {
	log     *log.Logger
	counter *Counter
}

func NewInit(l *log.Logger, c *Counter) *Init {
	return &Init{l, c}
}

func NewAbort(l *log.Logger, c *Counter) *Abort {
	return &Abort{l, c}
}

func (i *Init) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		i.log.Println("[INFO] Initializing:", i.counter.Me)

		m := Message{}
		if err := json.NewEncoder(rw).Encode(&m); err != nil {
			i.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		i.counter.acceptMessage(&m)
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *Abort) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.log.Println("[INFO] Aborting:", a.counter.Me)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			a.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		a.counter.abort(&m)
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
