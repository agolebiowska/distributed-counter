package main

type Counter struct {
    Items Items
}

type Item struct {
    ID     string `json:"id"`
    Tenant string `json:"tenant"`
}

type Items []*Item

func NewCounter(i Items) *Counter {
    return &Counter{
        Items: i,
    }
}