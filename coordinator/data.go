package main

import (
	"encoding/json"
	"io"
)

type Item struct {
	ID     string `json:"id"`
	Tenant string `json:"tenant"`
}

type Count struct {
	Value int `json:"count"`
}

type Items []*Item

func FromJSON(r io.Reader, i interface{}) error {
	d := json.NewDecoder(r)
	return d.Decode(&i)
}

func ToJSON(w io.Writer, i interface{}) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

// @todo: move to some kind of service
func GetItemsCountPerTenant(tenantID string) *Count {
	var count Count
	for _, v := range itemsList {
		if v.Tenant == tenantID {
			count.Value++
		}
	}

	return &count
}

var itemsList = []*Item{
	{
		ID:     "item-id-1",
		Tenant: "tenant-id-1",
	},
	{
		ID:     "item-id-2",
		Tenant: "tenant-id-2",
	},
	{
		ID:     "item-id-3",
		Tenant: "tenant-id-2",
	},
}
