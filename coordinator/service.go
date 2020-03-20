package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	u "github.com/agolebiowska/distributed-counter/utils"
)

type Counter struct {
	Addr     string
	HasItems bool
}

type Coordinator struct {
	Counters []Counter
}

// @todo: add required validation
type Item struct {
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
}

type Items []*Item

type Count struct {
	Value int `json:"count"`
}

type Message struct {
	ID      string `json:"id"`
	Content Items  `json:"content"`
}

func NewCounter(addr string) *Counter {
	return &Counter{
		Addr:     addr,
		HasItems: false,
	}
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		Counters: []Counter{},
	}
}

func NewMessage(items Items) *Message {
	return &Message{ID: time.Now().String(), Content: items}
}

func (c *Coordinator) acceptNewCounter(counterAddr string) {
	counter := NewCounter(counterAddr)
	c.Counters = append(c.Counters, *counter)
}

func (c *Coordinator) getItems() *Items {
	items := Items{}
	for _, ctr := range c.Counters {
		if !ctr.HasItems {
			continue
		}
		err := u.Do(http.MethodGet, fmt.Sprintf("http://%s", ctr.Addr), items, nil)
		if err != nil {
			log.Printf("[ERROR] Unable to get data from counter %s: %s", ctr.Addr, err.Error())
			continue
		}
	}
	return &items
}

func (c *Coordinator) getItemsCountPerTenant(tenantID string) *Count {
	count := Count{}
	err := u.Do(http.MethodGet, fmt.Sprintf("http://counter/%s", tenantID), count, nil)
	if err != nil {
		log.Printf("[ERROR] Unable to get data for tenant %s: %s", tenantID, err.Error())
	}
	return &count
}

func (c *Coordinator) canCommit(m *Message) bool {
	payload, _ := json.Marshal(m)
	// Sends CAN COMMIT to all counters
	for _, ctr := range c.Counters {
		err := u.Do(http.MethodPost, fmt.Sprintf("http://%s/init", ctr.Addr), nil, bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("[ERROR] Unable to init for %s: %s", ctr.Addr, err.Error())
			return false
		}
	}
	return true
}

func (c *Coordinator) abort(m *Message) error {
	payload, _ := json.Marshal(m)
	// Sends ABORT to all counters
	for _, ctr := range c.Counters {
		err := u.Do(http.MethodPost, fmt.Sprintf("http://%s/abort", ctr.Addr), nil, bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
	}
	return nil
}
