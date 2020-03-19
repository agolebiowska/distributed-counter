package utils

import (
    "bytes"
    "encoding/json"
    "io"
    "io/ioutil"
    "log"
    "net/http"
)

func Do(method string, urlStr string, result interface{}, buf *bytes.Buffer) error {
    req, err := http.NewRequest(method, urlStr, buf)
    if err != nil {
        return err
    }
    
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    defer func() {
        if resp != nil {
            resp.Body.Close()
        }
    }()
    if err != nil {
        return err
    }
    
    if resp.StatusCode != http.StatusOK {
        log.Printf("[ERROR] Unexpected status: %d %s\n", resp.StatusCode, err)
        return err
    }
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("[ERROR] Unable to read response body:", err)
        return err
    }
    
    err = json.Unmarshal(body, &result)
    if err != nil {
        log.Println("[ERROR] Unable to unmarshal json:", err)
        return err
    }
    
    return nil
}

func FromJSON(r io.Reader, i interface{}) error {
    d := json.NewDecoder(r)
    return d.Decode(&i)
}

func ToJSON(w io.Writer, i interface{}) error {
    e := json.NewEncoder(w)
    return e.Encode(i)
}