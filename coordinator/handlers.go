package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

type ItemsCount struct {
	coordinator *Coordinator
}

type ItemsAdd struct {
	coordinator *Coordinator
}

type CounterAdd struct {
	coordinator *Coordinator
}

type HealthCheck struct{}

type Status struct {
	Message string `json:"message"`
}

func NewItemsCount(c *Coordinator) *ItemsCount {
	return &ItemsCount{c}
}

func NewItemsAdd(c *Coordinator) *ItemsAdd {
	return &ItemsAdd{c}
}

func NewCounterAdd(c *Coordinator) *CounterAdd {
	return &CounterAdd{c}
}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func status(m string) string {
	j, _ := json.Marshal(&Status{Message: m})
	return string(j)
}

func (h *ItemsCount) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		l.Println("[INFO] Handle", r.Method, r.URL)
		rw.Header().Set("Content-Type", "application/json")

		// expect the tenant identifier in the URI
		reg := regexp.MustCompile(`\/items\/(.*)\/count`)
		g := reg.FindAllStringSubmatch(r.URL.Path, -1)
		if len(g) != 1 || len(g[0]) != 2 {
			l.Println("[ERROR] Invalid URI:", r.URL.Path)
			http.Error(rw, status("Invalid URI"), http.StatusBadRequest)
			return
		}

		count, err := h.coordinator.getItemsCountPerTenant(g[0][1])
		if err != nil {
			l.Println("[ERROR] Unable to get count:", err.Error())
			http.Error(rw, status("Unable to get count"), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(rw).Encode(count); err != nil {
			l.Println("[ERROR] Unable to marshall json:", err)
			http.Error(rw, status("Unable to marshall json"), http.StatusInternalServerError)
			return
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *ItemsAdd) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		l.Println("[INFO] Handle", r.Method, r.URL)
		rw.Header().Set("Content-Type", "application/json")

		items := Items{}
		if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
			l.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, status("Unable to unmarshal json"), http.StatusBadRequest)
			return
		}

		if err := items.Validate(); err != nil {
			l.Printf("[ERROR] Validation error: %s", err.Error())
			http.Error(rw, status(err.Error()), http.StatusBadRequest)
			return
		}

		m := NewMessage(items)
		if h.coordinator.canCommit(m) == false {
			h.coordinator.abort(m)
			http.Error(rw, status("Unable to add items"), http.StatusInternalServerError)
			return
		}

		if err := h.coordinator.commit(m); err != nil {
			http.Error(rw, status("Unable to add items"), http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(rw).Encode(Status{Message: "Success"}); err != nil {
			return
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *CounterAdd) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		l.Println("[INFO] Handle", r.Method, r.URL)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			l.Println("[ERROR] Unable to read body:", err.Error())
			http.Error(rw, status(fmt.Sprintf("Unable to read body: %s", err)), http.StatusBadRequest)
			return
		}

		items := h.coordinator.getItems()
		counterAddr := string(body)
		h.coordinator.acceptNewCounter(counterAddr)
		l.Println("[INFO] New counter accepted:", counterAddr)

		if err := json.NewEncoder(rw).Encode(items); err != nil {
			l.Println("[ERROR] Unable to marshal json:", err)
			http.Error(rw, status("Unable to marshal json"), http.StatusInternalServerError)
			return
		}

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HealthCheck) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	l.Println("[INFO] Health check")
}
