package main

import (
    "log"
    "net/http"
    
    u "github.com/agolebiowska/distributed-counter/utils"
)

type Counter struct {
    Addr     string
    HasItems bool
}

type Coordinator struct {
    Counters []Counter
}

type Item struct {
    ID     string `json:"id"`
    Tenant string `json:"tenant"`
}

type Items []*Item

type Count struct {
    Value int `json:"count"`
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

func (c *Coordinator) GetItems() *Items {
    items := Items{}
    for _, counter := range c.Counters {
        if !counter.HasItems {
            continue
        }
        err := u.Do(http.MethodGet, counter.Addr, items, nil)
        if err != nil {
            log.Printf("[ERROR]: Unable to get data from counter %s: %s", counter.Addr, err.Error())
            continue
        }
    }
    return &items
}

func (c *Coordinator) GetItemsCountPerTenant(tenantID string) *Count {
    count := Count{}
    err := u.Do(http.MethodGet, "hostname container or something/"+tenantID, count, nil)
    if err != nil {
        log.Printf("[ERROR]: Unable to get data from counter %s: %s", "hostname container or something/"+tenantID, err.Error())
    }
    return &count
}