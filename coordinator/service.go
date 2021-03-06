package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type Counter struct {
	Addr          string
	HasItems      bool
	IsDead        bool
	RecoveryTries int16
}

type Coordinator struct {
	Counters []*Counter

	http *http.Client
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
		IsDead:   false,
	}
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		Counters: []*Counter{},

		http: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

func NewMessage(items Items) *Message {
	return &Message{ID: uuid(), Content: items}
}

// naive uuid, do not use in production
// @todo: add uuid package
func uuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return time.Now().String()
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (c *Coordinator) acceptNewCounter(counterAddr string) {
	counter := NewCounter(counterAddr)
	c.Counters = append(c.Counters, counter)
}

// sends GET request to alive and populated counter
// returns all items
func (c *Coordinator) getItems() Items {
	items := Items{}
	var body []byte
	for _, counter := range c.Counters {
		if counter.IsDead || counter.HasItems == false {
			continue
		}

		resp, err := c.Do(http.MethodGet, fmt.Sprintf("http://%s/items", counter.Addr), nil)
		if err != nil {
			l.Printf("[ERROR] Cannot get items from counter %s: %s", counter.Addr, err.Error())
			continue
		}

		if resp.StatusCode != http.StatusOK {
			l.Printf("[ERROR] Unexpected status code %d from %s", resp.StatusCode, counter.Addr)
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Printf("[ERROR] Cannot read response body from %s: %s", counter.Addr, err.Error())
			continue
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if body != nil {
		err := json.Unmarshal(body, &items)
		if err != nil {
			l.Printf("[ERROR] Cannot unmarshal json: %s ", err.Error())
		}
	}

	return items
}

// sends GET request to random counter
// returns counted items for given tenantID
func (c *Coordinator) getItemsCountPerTenant(tenantID string) (*Count, error) {
	count := Count{}

	url := fmt.Sprintf("http://counter/items/%s/count", tenantID)
	resp, err := c.Do(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		l.Printf("[ERROR] Unexpected status code: %d", resp.StatusCode)
		return nil, errors.New("unexpected status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Printf("[ERROR] Cannot read response body: %s", err.Error())
		return nil, err
	}

	err = json.Unmarshal(body, &count)
	if err != nil {
		l.Printf("[ERROR] Cannot unmarshal json: %s ", err.Error())
		return nil, err
	}

	return &count, nil
}

// sends POST request to every counter
// returns information whether all counters are ready to save data
func (c *Coordinator) canCommit(m *Message) bool {
	payload, err := json.Marshal(m)
	if err != nil {
		l.Printf("[ERROR] Unable to marshall message %+v: %s", m, err.Error())
	}

	checked := 0
	agrees := make([]bool, 0)
	for _, counter := range c.Counters {
		if counter.IsDead {
			continue
		}

		url := fmt.Sprintf("http://%s/init", counter.Addr)
		resp, err := c.Do(http.MethodPost, url, bytes.NewBuffer(payload))
		if err != nil {
			l.Printf("[ERROR] Cannot init for %s: %s", counter.Addr, err.Error())
			continue
		}

		if resp.StatusCode == http.StatusOK {
			agrees = append(agrees, true)
		}

		checked++
		if resp != nil {
			resp.Body.Close()
		}
	}

	return len(agrees) == checked
}

// sends POST request to every counter
// to delete a previously initiated message
func (c *Coordinator) abort(m *Message) {
	payload, err := json.Marshal(m)
	if err != nil {
		l.Printf("[ERROR] Unable to marshall message %+v: %s", m, err.Error())
	}

	for _, counter := range c.Counters {
		if counter.IsDead {
			continue
		}

		url := fmt.Sprintf("http://%s/abort", counter.Addr)
		resp, err := c.Do(http.MethodPost, url, bytes.NewBuffer(payload))
		if err != nil {
			l.Printf("[ERROR] Unable to abort %s: %s", counter.Addr, err.Error())
			return
		}

		if resp.StatusCode != http.StatusOK {
			l.Printf("[ERROR] Unable to abort %s: %d", counter.Addr, resp.StatusCode)
			return
		}

		if resp != nil {
			resp.Body.Close()
		}
	}
}

// sends POST request to every counter
// to save data from previously initiated message
func (c *Coordinator) commit(m *Message) error {
	payload, err := json.Marshal(m)
	if err != nil {
		l.Printf("[ERROR] Unable to marshall message %+v: %s", m, err.Error())
	}

	for _, counter := range c.Counters {
		if counter.IsDead {
			continue
		}

		url := fmt.Sprintf("http://%s/commit", counter.Addr)
		resp, err := c.Do(http.MethodPost, url, bytes.NewBuffer(payload))
		if err != nil {
			l.Printf("[ERROR] Unable to commit %s: %s", counter.Addr, err.Error())
			return err
		}

		if resp.StatusCode != http.StatusOK {
			l.Printf("[ERROR] Unable to commit %s: %d", counter.Addr, resp.StatusCode)
			return errors.New("unable to commit")
		}

		counter.HasItems = true
		if resp != nil {
			resp.Body.Close()
		}
	}
	return nil
}

func (c *Coordinator) Do(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	return resp, err
}
