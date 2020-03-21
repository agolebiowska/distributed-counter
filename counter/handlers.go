package main

import (
	"encoding/json"
	"net/http"
	"regexp"
)

type Init struct {
	counter *Counter
}

type Abort struct {
	counter *Counter
}

type Commit struct {
	counter *Counter
}

type CountItems struct {
	counter *Counter
}

type ItemsGet struct {
	counter *Counter
}

type HealthCheck struct {
	counter *Counter
}

func NewInit(c *Counter) *Init {
	return &Init{c}
}

func NewAbort(c *Counter) *Abort {
	return &Abort{c}
}

func NewCommit(c *Counter) *Commit {
	return &Commit{c}
}

func NewCountItems(c *Counter) *CountItems {
	return &CountItems{c}
}

func NewItemsGet(c *Counter) *ItemsGet {
	return &ItemsGet{c}
}

func NewHealthCheck(c *Counter) *HealthCheck {
	return &HealthCheck{c}
}

func (h *CountItems) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		l.Println("[INFO] Handle", r.Method, r.URL)
		rw.Header().Set("Content-Type", "application/json")

		// expect the tenant identifier in the URI
		reg := regexp.MustCompile(`\/items\/(.*)\/count`)
		g := reg.FindAllStringSubmatch(r.URL.Path, -1)
		if len(g) != 1 || len(g[0]) != 2 {
			l.Println("[ERROR] Invalid URI:", r.URL.Path)
			http.Error(rw, "Invalid URI", http.StatusBadRequest)
			return
		}

		tenantID := g[0][1]
		count := h.counter.countItemsForTenant(tenantID)
		if err := json.NewEncoder(rw).Encode(count); err != nil {
			l.Println("[ERROR] Unable to marshall json:", err)
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
		l.Println("[INFO] Handle", r.Method, r.URL)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			l.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		h.counter.acceptMessage(&m)
		l.Printf("[INFO] %s initialized: %+v", h.counter.Me, m)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Abort) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		l.Println("[INFO] Handle", r.Method, r.URL)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			l.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		h.counter.abort(&m)
		l.Printf("[INFO] %s aborted: %+v", h.counter.Me, m)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Commit) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		l.Println("[INFO] Handle", r.Method, r.URL)

		m := Message{}
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			l.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		h.counter.commit(&m)
		l.Printf("[INFO] %s committed: %+v", h.counter.Me, m)

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ItemsGet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		l.Println("[INFO] Handle", r.Method, r.URL)
		rw.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(rw).Encode(h.counter.Items); err != nil {
			l.Println("[ERROR] Unable to marshall json:", err)
			http.Error(rw, "Unable to marshall json", http.StatusInternalServerError)
			return
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HealthCheck) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	l.Printf("[INFO] %s healthy", h.counter.Me)
}
