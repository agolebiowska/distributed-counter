package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Counter struct {
	Addr     string
	HasItems bool
}

type Coordinator struct {
	Counters []*Counter
}

type Item struct {
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
}

type Items []Item

type Count struct {
	Value int `json:"count"`
}

type Message struct {
	ID      string `json:"id"`
	Content Items  `json:"content"`
}

func (i *Items) Validate() error {
	for _, v := range *i {
		if v.ID == "" || v.Tenant == "" {
			return errors.New("both values are required")
		}
	}
	return nil
}

func NewCounter(addr string) *Counter {
	return &Counter{
		Addr:     addr,
		HasItems: false,
	}
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		Counters: []*Counter{},
	}
}

func NewMessage(items Items) *Message {
	return &Message{ID: time.Now().String(), Content: items}
}

func (c *Coordinator) acceptNewCounter(counterAddr string) {
	counter := NewCounter(counterAddr)
	c.Counters = append(c.Counters, counter)
}

func (c *Coordinator) getItems() Items {
	items := Items{}
	var body []byte
	for _, counter := range c.Counters {
		if counter.HasItems == false {
			continue
		}

		resp, err := Do(http.MethodGet, fmt.Sprintf("http://%s/items", counter.Addr), nil)
		if err != nil {
			log.Printf("[ERROR] Cannot get items from counter %s: %s", counter.Addr, err.Error())
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[ERROR] Unexpected status code %d from %s", resp.StatusCode, counter.Addr)
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[ERROR] Cannot read response body from %s: %s", counter.Addr, err.Error())
			continue
		}
		counter.HasItems = true
	}

	if body != nil {
		err := json.Unmarshal(body, &items)
		if err != nil {
			log.Printf("[ERROR] Cannot unmarshal %s: ", err.Error())
		}
	}

	return items
}

func (c *Coordinator) getItemsCountPerTenant(tenantID string) (*Count, error) {
	count := Count{}
	var body []byte
	url := fmt.Sprintf("http://counter/items/%s/count", tenantID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Cannot read response body: %s", err.Error())
		return nil, err
	}

	err = json.Unmarshal(body, &count)
	if err != nil {
		log.Printf("[ERROR] Cannot unmarshal %s: ", err.Error())
		return nil, err
	}

	return &count, nil
}

func (c *Coordinator) canCommit(m *Message) bool {
	payload, err := json.Marshal(m)
	if err != nil {
		log.Printf("[ERROR] Unable to marshall message %+v: %s", m, err.Error())
	}

	agrees := make([]bool, 0)
	for _, counter := range c.Counters {
		url := fmt.Sprintf("http://%s/init", counter.Addr)
		resp, err := Do(http.MethodPost, url, bytes.NewBuffer(payload))
		defer func(resp *http.Response) {
			if resp != nil {
				resp.Body.Close()
			}
		}(resp)
		if err != nil {
			log.Printf("[ERROR] Cannot init for %s: %s", counter.Addr, err.Error())
		}

		if resp.StatusCode == http.StatusOK {
			agrees = append(agrees, true)
		}
	}

	return len(agrees) == len(c.Counters)
}

func (c *Coordinator) abort(m *Message) {
	payload, err := json.Marshal(m)
	if err != nil {
		log.Printf("[ERROR] Unable to marshall message %+v: %s", m, err.Error())
	}

	for _, counter := range c.Counters {
		url := fmt.Sprintf("http://%s/abort", counter.Addr)
		resp, err := Do(http.MethodPost, url, bytes.NewBuffer(payload))
		defer func(resp *http.Response) {
			if resp != nil {
				resp.Body.Close()
			}
		}(resp)
		if err != nil {
			log.Printf("[ERROR] Unable to abort %s: %s", counter.Addr, err.Error())
			return
		}
	}
}

func (c *Coordinator) commit(m *Message) {
	payload, err := json.Marshal(m)
	if err != nil {
		log.Printf("[ERROR] Unable to marshall message %+v: %s", m, err.Error())
	}

	for _, counter := range c.Counters {
		url := fmt.Sprintf("http://%s/commit", counter.Addr)
		resp, err := Do(http.MethodPost, url, bytes.NewBuffer(payload))
		defer func(resp *http.Response) {
			if resp != nil {
				resp.Body.Close()
			}
		}(resp)
		if err != nil {
			log.Printf("[ERROR] Unable to commit %s: %s", counter.Addr, err.Error())
			return
		}
		counter.HasItems = true
	}
}
