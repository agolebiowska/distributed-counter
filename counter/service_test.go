package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestCounter_SignIn(t *testing.T) {
	items := Items{
		{
			ID:     "item-1",
			Tenant: "test",
		},
		{
			ID:     "item-2",
			Tenant: "test",
		},
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`[{"id":"item-1","tenant":"test"}, {"id":"item-2","tenant":"test"}]`)),
			Header:     make(http.Header),
		}
	})

	c := &Counter{
		Me:   "counter",
		http: client,
	}

	if len(c.Items) > 0 {
		t.Error("New counter must have empty items")
	}

	if err := c.SignIn(); err != nil {
		t.Errorf("SignIn error: %s", err.Error())
	}

	if !reflect.DeepEqual(items, c.Items) {
		t.Errorf("Want %+v, got %+v", items, c.Items)
	}
}
