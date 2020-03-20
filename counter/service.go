package main

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

type Items []*Item
type Messages []*Message

func NewCounter(m string, i Items) *Counter {
	return &Counter{
		Me:    m,
		Items: i,
	}
}

func (c *Counter) acceptMessage(m *Message) {
	c.Messages = append(c.Messages, m)
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
