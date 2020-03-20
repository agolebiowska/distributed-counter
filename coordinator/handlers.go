package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type ItemsCount struct {
	log         *log.Logger
	coordinator *Coordinator
}

type ItemsAdd struct {
	log         *log.Logger
	coordinator *Coordinator
}

type CounterAdd struct {
	log         *log.Logger
	coordinator *Coordinator
}

func NewItemsCount(l *log.Logger, c *Coordinator) *ItemsCount {
	return &ItemsCount{l, c}
}

func NewItemsAdd(l *log.Logger, c *Coordinator) *ItemsAdd {
	return &ItemsAdd{l, c}
}

func NewCounterAdd(l *log.Logger, c *Coordinator) *CounterAdd {
	return &CounterAdd{l, c}
}

func (i *ItemsCount) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// expect the tenant identifier in the URI
		reg := regexp.MustCompile(`\/items\/(.*)\/count`)
		g := reg.FindAllStringSubmatch(r.URL.Path, -1)
		if len(g) != 1 || len(g[0]) != 2 {
			i.log.Println("[ERROR] Invalid URI:", r.URL.Path)
			http.Error(rw, "Invalid URI", http.StatusBadRequest)
			return
		}

		tenantID := g[0][1]
		i.log.Println("[INFO] Handle GET items count per tenant:", tenantID)

		count := i.coordinator.getItemsCountPerTenant(tenantID)
		if err := json.NewEncoder(rw).Encode(&count); err != nil {
			i.log.Println("[ERROR] Unable to marshal json:", err)
			http.Error(rw, "Unable to marshall json", http.StatusInternalServerError)
		}
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (i *ItemsAdd) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		i.log.Println("[INFO] Handle POST items")

		items := Items{}
		if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
			i.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		m := NewMessage(items)
		if i.coordinator.canCommit(m) == false {
			err := i.coordinator.abort(m)
			if err != nil {
				i.log.Println("[ERROR] Unable to abort:", err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte("Unable to abort"))
			}
		}
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *CounterAdd) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.log.Println("[INFO] Handle POST counter")

		counterAddrByte, err := ioutil.ReadAll(r.Body)
		if err != nil {
			c.log.Println("[ERROR] Unable to read body:", err)
			http.Error(rw, "Unable to read body", http.StatusBadRequest)
			return
		}
		counterAddr := string(counterAddrByte)
		items := c.coordinator.getItems()
		c.coordinator.acceptNewCounter(counterAddr)

		if err := json.NewEncoder(rw).Encode(&items); err != nil {
			log.Println("[ERROR] Unable to marshal json:", err)
			http.Error(rw, "Unable to marshall json", http.StatusInternalServerError)
			return
		}
		c.log.Println("[INFO] New counter accepted:", counterAddr)
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
