/*
This is my first Go program so probably a few things are not made in the
best way.
*/

package main

import (
    "os"
    "log"
    "io/ioutil"
    "math/rand"
    "strconv"
    "net/http"
    "encoding/json"
)

var ALLOWED_ORIGIN = os.Getenv("ALLOWED_ORIGIN")
var CLIENT_DATA = make(map[string]*Data)


type Dimension struct {
    Width  string
    Height string
}

type Data struct {
    WebsiteUrl         string
    SessionId          string
    ResizeFrom         Dimension
    ResizeTo           Dimension
    CopyAndPaste       map[string]bool // map[fieldId]true
    FormCompletionTime int `json:"time"`// Seconds
}

type JSONEvent struct {
    EventType          string
}

type JSONPasteEvent struct {
    *JSONEvent
    Pasted             bool
    FormId             string
}


// Handle payload from a client
func payloadHandler(w http.ResponseWriter, r *http.Request, sessionId string) {

    if _, ok := CLIENT_DATA[sessionId]; ! ok {
        CLIENT_DATA[sessionId] = &Data{ CopyAndPaste: make(map[string]bool)}
    }

    // Probably inefficient. We use a buffer in order to be able and
    // unmarshal the payload to two different structs. The memory expense
    // for simple payloads is neglitible but there might a better way..
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Printf("Could not read body content: %s", err)
        http.Error(w, "Error reading request body", http.StatusInternalServerError)
    }

    err = json.Unmarshal(body, CLIENT_DATA[sessionId])
    if err != nil {
        log.Printf("Could not unmarshal Data: %s", err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
    }

    // Unmarshaling paste event needs a few more steps
    var event JSONPasteEvent
    err = json.Unmarshal(body, &event)
    if err != nil {
        log.Printf("Could not unmarshal PasteEvent: %s", err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
    }
    if event.EventType == "copyAndPaste" {
        CLIENT_DATA[sessionId].CopyAndPaste[event.FormId] = event.Pasted
    }

    log.Printf("%+v\n", *CLIENT_DATA[sessionId])
}


// Handle client request
func clientHandler(w http.ResponseWriter, r *http.Request) {

    if r.Method == "POST" {

        // Allow ONLY proper json
        contentType := r.Header.Get("Content-type")
        if contentType != "application/json" {
            log.Printf("Not accepting content type '%s'\n", contentType)
            http.Error(w, "Request content not supported", http.StatusNotAcceptable)
        }

        // Set session ID for client if needed
        var sessionId string
        cookie, _ := r.Cookie("sessionId")
        if cookie == nil {
            sessionId = strconv.Itoa(rand.Intn(100))
            newCookie := http.Cookie{Name: "sessionId", Value:sessionId}
            http.SetCookie(w, &newCookie)
            log.Printf("Set new sessionId: %s\n", sessionId)
        } else {
            sessionId = cookie.Value
        }

        w.Header().Set("Access-Control-Allow-Origin", ALLOWED_ORIGIN)
        w.Header().Set("Access-Control-Allow-Credentials", "true")

        payloadHandler(w, r, sessionId)
    } else if r.Method == "OPTIONS" { // Deal with CORS preflight
        w.Header().Set("Access-Control-Allow-Origin", ALLOWED_ORIGIN)
        w.Header().Set("Access-Control-Allow-Credentials", "true")
        w.Header().Set("Access-Control-Allow-Methods", "POST")
        w.Header().Set("Access-Control-Max-Age", "1000")
        w.Header().Set("Access-Control-Allow-Headers", "origin, x-csrf-token, content-type, accept")
    } else {
        log.Printf("Omitting method '%s'\n", r.Method)
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}


func main() {
    http.HandleFunc("/", clientHandler)
    http.ListenAndServe(":8080", nil)
}
