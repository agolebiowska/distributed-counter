package main

import (
    "bytes"
    "log"
    "net/http"
    "os"
    
    u "github.com/agolebiowska/distributed-counter/utils"
)

func main() {
	// @todo: "apply" to coordinator with our address which gets us all items?
	me, err := os.Hostname()
	if err != nil {
	    log.Fatal("[ERROR] Cannot obtain hostname:", err.Error())
    }
    
    cAddr := os.Getenv("COORDINATOR_ADDR")
    
    data := []byte(me)
    items := Items{}
    err = u.Do(http.MethodPost, cAddr, items, bytes.NewBuffer(data))
    if err != nil {
        log.Fatal("[ERROR]: Cannot add counter:", err.Error())
    }
    
    c := NewCounter(items)
    
	// if ok then start the server
}
