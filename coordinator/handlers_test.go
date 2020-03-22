package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// returns *http.Client with Transport replaced
// to avoid making real calls to counters
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func resp(code int) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(bytes.NewBufferString(`KO`)),
		Header:     make(http.Header),
	}
}

func TestItemsAdd_ServeHTTP(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		switch req.URL.Host {
		case "noError":
			return resp(200)
		case "initError":
			return initError(req.URL.Path)
		case "abortError":
			return abortError(req.URL.Path)
		case "commitError":
			return commitError(req.URL.Path)
		default:
			return resp(500)
		}
	})

	tt := []struct {
		name       string
		method     string
		counters   []*Counter
		body       string
		want       string
		statusCode int
	}{
		{
			name:       "wrong HTTP method",
			method:     http.MethodGet,
			counters:   []*Counter{},
			body:       `[{"ID":"item-1", "tenant":"tenant-1"}]`,
			want:       ``,
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "invalid body",
			method:     http.MethodPost,
			counters:   []*Counter{},
			body:       `[{"ID":"", "tenant":""}]`,
			want:       `Validation error: both values are required`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "counter init fail",
			method: http.MethodPost,
			counters: []*Counter{
				{Addr: "noError", HasItems: true},
				{Addr: "noError", HasItems: true},
				{Addr: "initError", HasItems: true},
			},
			body:       `[{"ID":"item-1", "tenant":"tenant-1"}]`,
			want:       `Unable to add items`,
			statusCode: http.StatusInternalServerError,
		},
		{
			name:   "counter commit fail",
			method: http.MethodPost,
			counters: []*Counter{
				{Addr: "noError", HasItems: true},
				{Addr: "noError", HasItems: true},
				{Addr: "commitError", HasItems: true},
			},
			body:       `[{"ID":"item-1", "tenant":"tenant-1"}]`,
			want:       `{"message":"Success"}`,
			statusCode: http.StatusOK,
		},
		{
			name:   "no counter fail",
			method: http.MethodPost,
			counters: []*Counter{
				{Addr: "noError", HasItems: true},
				{Addr: "noError", HasItems: true},
				{Addr: "noError", HasItems: true},
			},
			body:       `[{"ID":"item-1", "tenant":"tenant-1"}]`,
			want:       `{"message":"Success"}`,
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/items", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()

			c := &Coordinator{
				Counters: tc.counters,
				http:     client,
			}
			NewItemsAdd(c).ServeHTTP(rr, request)

			if rr.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, rr.Code)
			}

			if strings.TrimSpace(rr.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, rr.Body)
			}
		})
	}
}

func TestHealthCheck_ServeHTTP(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		want       string
		statusCode int
	}{
		{
			name:       "health check",
			method:     http.MethodGet,
			want:       "",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/health", nil)
			rr := httptest.NewRecorder()

			NewHealthCheck().ServeHTTP(rr, request)

			if rr.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, rr.Code)
			}

			if strings.TrimSpace(rr.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'", tc.want, rr.Body)
			}
		})
	}
}

func initError(p string) *http.Response {
	switch p {
	case "/init":
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`KO`)),
			Header:     make(http.Header),
		}
	case "/abort":
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	default:
		return resp(500)
	}
}

func abortError(p string) *http.Response {
	switch p {
	case "/init":
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`KO`)),
			Header:     make(http.Header),
		}
	case "/abort":
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	default:
		return resp(500)
	}
}

func commitError(p string) *http.Response {
	switch p {
	case "/init":
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`KO`)),
			Header:     make(http.Header),
		}
	case "/commit":
		return &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
			Header:     make(http.Header),
		}
	default:
		return resp(500)
	}
}
