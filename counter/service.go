package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var coordinatorAddr = "http://coordinator"

type Counter struct {
	Me       string
	Items    Items
	Messages Messages

	http *http.Client
}

type Item struct {
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
}

type Message struct {
	ID      string `json:"id"`
	Content Items  `json:"content"`
}

type Count struct {
	Value int `json:"count"`
}

type Items []Item
type Messages []Message

func NewCounter(m string) *Counter {
	return &Counter{
		Me: m,

		http: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

func (c *Counter) countItemsForTenant(tenantID string) *Count {
	items := map[string]bool{}
	for _, i := range c.Items {
		if i.Tenant == tenantID {
			items[i.ID] = true
		}
	}
	return &Count{Value: len(items)}
}

func (c *Counter) acceptMessage(m *Message) {
	c.Messages = append(c.Messages, *m)
}

func (c *Counter) abort(m *Message) {
	for i, mess := range c.Messages {
		if mess.ID == m.ID {
			c.Messages = append(c.Messages[:i], c.Messages[i+1:]...)
			break
		}
	}
}

func (c *Counter) commit(m *Message) {
	for i, mess := range c.Messages {
		if mess.ID == m.ID {
			c.Items = append(c.Items, m.Content...)
			c.Messages = append(c.Messages[:i], c.Messages[i+1:]...)
			break
		}
	}
}

func (c *Counter) SignIn() error {
	myAddr := []byte(c.Me)
	url := fmt.Sprintf("%s/counters", coordinatorAddr)
	resp, err := c.Do(http.MethodPost, url, bytes.NewBuffer(myAddr))
	defer func(resp *http.Response) {
		if resp != nil {
			resp.Body.Close()
		}
	}(resp)
	if err != nil {
		l.Printf("[ERROR] Add counter error: %s", err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		l.Printf("[ERROR] Unexpected status code %d for add counter: %s", resp.StatusCode, err)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Printf("[ERROR] Cannot read from add counter: %s", err.Error())
		return err
	}

	items := Items{}
	if err := json.Unmarshal(body, &items); err != nil {
		l.Printf("[ERROR] Cannot unmarshall json: %s", body)
		return err
	}
	c.Items = items

	return nil
}

func (c *Counter) Do(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	return resp, err
}
