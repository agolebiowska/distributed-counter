package main

type Counter struct {
	Addr     string
	HasItems bool
}

type Coordinator struct {
	Counters []Counter
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

func (c *Coordinator) AcceptNewCounter(counterAddr string) {
	counter := NewCounter(counterAddr)
	c.Counters = append(c.Counters, *counter)
}
