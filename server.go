package locust

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Server datastructure encapsolating data for connections
// and targets
type Server struct {
	addr    string
	clients map[string]clientConn
	targets map[string]resultSet
	showUI  bool
}

type clientConn struct {
	name    string
	addr    string
	actions chan action
}

// NewServer creates a new instance of a server
func NewServer(addr string, ui bool) *Server {
	return &Server{
		addr:    addr,
		clients: make(map[string]clientConn),
		targets: map[string]resultSet{},
		showUI:  ui,
	}
}

// Start Listening on http for incoming client connections
func (s *Server) Start() error {
	http.HandleFunc("/join", s.join)
	http.HandleFunc("/request", s.request)
	http.HandleFunc("/report", s.report)
	http.HandleFunc("/results", s.results)
	if s.showUI {
		http.Handle("/_/", http.StripPrefix("/_/", http.FileServer(http.Dir("frontend/dist"))))
		http.HandleFunc("/", s.serveWebUI)
	}

	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) join(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Flusher api not available for this connection", 400)
		return
	}
	name := r.Header.Get("client-name")
	actions := make(chan action)
	s.clients[name] = clientConn{
		name:    name,
		addr:    r.RemoteAddr,
		actions: actions,
	}
	enc := json.NewEncoder(w)
	for action := range actions {
		err := enc.Encode(&action)
		if err != nil {
			log.Println("Error encoding action", err)
			continue
		}
		flusher.Flush()
	}
}

func (s *Server) request(w http.ResponseWriter, r *http.Request) {
	requestID := "test-request"
	target := r.URL.Query().Get("target")
	log.Printf("Request received to target %v sending to %v\n", target, s.clients)
	s.targets[requestID] = resultSet{
		count:   len(s.clients),
		results: map[string]int{},
	}
	for _, client := range s.clients {
		client.actions <- action{
			RequestID: requestID,
			Type:      ping,
			Target:    target,
		}
	}
	fmt.Fprint(w, requestID)
}

func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	var res response
	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid response", http.StatusBadRequest)
		return
	}
	s.targets[res.RequestID].results[res.ClientName] = res.Ping
	log.Printf("Report received %v\n", s.targets)
}

func (s *Server) results(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Flusher api not available for this connection", 400)
		return
	}
	requestID := r.URL.Query().Get("requestID")
	enc := json.NewEncoder(w)
	// TODO: fix me pls owo
	for {
		res := s.targets[requestID]
		err := enc.Encode(&res.results)
		if err != nil {
			break
		}
		flusher.Flush()
		if len(res.results) == res.count {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (s *Server) serveWebUI(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<title>Locust</title>
		<div id="main"></div>
		<script src="/_/main.js"></script>
	`)
}
