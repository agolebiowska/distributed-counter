package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type ItemsCount struct {
	log *log.Logger
}

type ItemsAdd struct {
	log *log.Logger
}

type CounterAdd struct {
	log         *log.Logger
	coordinator *Coordinator
}

func NewItemsCount(l *log.Logger) *ItemsCount {
	return &ItemsCount{l}
}

func NewItemsAdd(l *log.Logger) *ItemsAdd {
	return &ItemsAdd{l}
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

		count := GetItemsCountPerTenant(tenantID)
		err := ToJSON(rw, count)
		if err != nil {
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

		err := FromJSON(r.Body, Items{})
		if err != nil {
			i.log.Println("[ERROR] Unable to unmarshal json:", err)
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		// @todo: handle counter communication etc
		rw.Write([]byte("items added"))
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
		c.coordinator.AcceptNewCounter(counterAddr)
		c.log.Println("[INFO] New counter accepted:", counterAddr)
		rw.Write([]byte("New counter accepted"))
		return

	default:
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
