package handlers

import (
    "log"
    "net/http"
    "regexp"
    
    "coordinator/data"
)

type ItemsCount struct {
    l *log.Logger
}

type ItemsAdd struct {
    l *log.Logger
}

func NewItemsCount(l *log.Logger) *ItemsCount {
    return &ItemsCount{l}
}

func NewItemsAdd(l *log.Logger) *ItemsAdd {
    return &ItemsAdd{l}
}

func (i *ItemsCount) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // expect the tenant identifier in the URI
        reg := regexp.MustCompile(`\/items\/(.*)\/count`)
        g := reg.FindAllStringSubmatch(r.URL.Path, -1)
        if len(g) != 1 || len(g[0]) != 2 {
            i.l.Println("[ERROR] Invalid URI:", r.URL.Path)
            http.Error(rw, "Invalid URI", http.StatusBadRequest)
            return
        }
        
        i.GET(g[0][1], rw)
        return
    
    default:
        rw.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func (i *ItemsAdd) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        i.POST(rw, r)
        return
    
    default:
        rw.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func (i *ItemsCount) GET(tenantID string, rw http.ResponseWriter) {
    i.l.Println("[INFO] Handle GET items count per tenant:", tenantID)
    
    count := data.GetItemsCountPerTenant(tenantID)
    err := count.ToJSON(rw)
    if err != nil {
        i.l.Println("[ERROR] Unable to marshal json:", err)
        http.Error(rw, "Unable to marshall json", http.StatusInternalServerError)
        return
    }
}

func (i *ItemsAdd) POST(rw http.ResponseWriter, r *http.Request) {
    i.l.Println("[INFO] Handle POST items")
    
    items := data.Items{}
    err := items.FromJSON(r.Body)
    if err != nil {
        i.l.Println("[ERROR] Unable to unmarshal json:", err)
        http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
        return
    }
    
    // @todo: handle counter communication etc
    rw.Write([]byte("items added"))
}
