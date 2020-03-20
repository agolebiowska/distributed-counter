package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

type Init struct {
	log     *log.Logger
	counter *Counter
}

type Abort struct {
	log     *log.Logger
	counter *Counter
}

type Commit struct {
	log     *log.Logger
	counter *Counter
}

type CountItems struct {
	log     *log.Logger
	counter *Counter
}

func NewInit(l *log.Logger, c *Counter) *Init {
	return &Init{l, c}
}

func NewAbort(l *log.Logger, c *Counter) *Abort {
	return &Abort{l, c}
}

func NewCommit(l *log.Logger, c *Counter) *Commit {
	return &Commit{l, c}
}

func NewCountItems(l *log.Logger, c *Counter) *CountItems {
	return &CountItems{l, c}
}

func (c *CountItems) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.log.Println("[INFO] Counting items:", c.counter.Me)

		// expect the tenant identifier in the URI
		reg := regexp.MustCompile(`\/items\/(.*)\/count`)
		g := reg.FindAllStringSubmatch(r.URL.Path, -1)
		if len(g) != 1 || len(g[0]) != 2 {
			c.log.Println("[ERROR] Invalid URI:", r.URL.Path)
			http.Error(rw, "Invalid URI", http.StatusBadRequest)
			return
		}

		tenantID := g[0][1]
		count := c.counter.countItemsForTenant(tenantID)
		if err := json.NewEncoder(rw).Encode(&count); err != nil {
			c.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
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

func (c *Commit) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.log.Println("[INFO] Committing:", c.counter.Me)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			c.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		c.counter.commit(&m)
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
