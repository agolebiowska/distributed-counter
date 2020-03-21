package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var coordinatorAddr = "http://coordinator"

type Counter struct {
	Me       string
	Items    Items
	Messages Messages
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

func NewCounter(m string, i Items) *Counter {
	return &Counter{
		Me:    m,
		Items: i,
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

func SignIn(me string) (Items, error) {
	myAddr := []byte(me)
	url := fmt.Sprintf("%s/counters", coordinatorAddr)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(myAddr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	defer func(resp *http.Response) {
		if resp != nil {
			resp.Body.Close()
		}
	}(resp)
	if err != nil {
		l.Printf("[ERROR] Add counter error: %s", err.Error())
		return Items{}, err
	}

	if resp.StatusCode != http.StatusOK {
		l.Printf("[ERROR] Unexpected status code %d for add counter: %s", resp.StatusCode, err)
		return Items{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Printf("[ERROR] Cannot read from add counter: %s", err.Error())
		return Items{}, err
	}

	items := Items{}
	if err := json.Unmarshal(body, &items); err != nil {
		l.Printf("[ERROR] Cannot unmarshall json: %s", body)
		return Items{}, err
	}

	return items, nil
}
