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

type ItemsGet struct {
	log     *log.Logger
	counter *Counter
}

type HealthCheck struct {
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

func NewItemsGet(l *log.Logger, c *Counter) *ItemsGet {
	return &ItemsGet{l, c}
}

func NewHealthCheck(l *log.Logger, c *Counter) *HealthCheck {
	return &HealthCheck{l, c}
}

func (h *CountItems) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.log.Println("[INFO] Handle", r.Method, r.URL)
		rw.Header().Set("Content-Type", "application/json")

		// expect the tenant identifier in the URI
		reg := regexp.MustCompile(`\/items\/(.*)\/count`)
		g := reg.FindAllStringSubmatch(r.URL.Path, -1)
		if len(g) != 1 || len(g[0]) != 2 {
			h.log.Println("[ERROR] Invalid URI:", r.URL.Path)
			http.Error(rw, "Invalid URI", http.StatusBadRequest)
			return
		}

		tenantID := g[0][1]
		count := h.counter.countItemsForTenant(tenantID)
		if err := json.NewEncoder(rw).Encode(count); err != nil {
			h.log.Println("[ERROR] Unable to marshall json:", err)
			http.Error(rw, "Unable to marshall json", http.StatusInternalServerError)
			return
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Init) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.log.Println("[INFO] Handle", r.Method, r.URL)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			h.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		h.counter.acceptMessage(&m)
		h.log.Printf("[INFO] %s initialized: %+v", h.counter.Me, m)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Abort) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.log.Println("[INFO] Handle", r.Method, r.URL)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			h.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		h.counter.abort(&m)
		h.log.Printf("[INFO] %s aborted: %+v", h.counter.Me, m)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Commit) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.log.Println("[INFO] Handle", r.Method, r.URL)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			h.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		h.counter.commit(&m)
		h.log.Printf("[INFO] %s committed: %+v", h.counter.Me, m)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ItemsGet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.log.Println("[INFO] Handle", r.Method, r.URL)
		rw.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(rw).Encode(h.counter.Items); err != nil {
			h.log.Println("[ERROR] Unable to marshall json:", err)
			http.Error(rw, "Unable to marshall json", http.StatusInternalServerError)
			return
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HealthCheck) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.log.Printf("[INFO] %s healthy", h.counter.Me)
}
