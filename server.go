package locust

import (
	"encoding/json"
	"log"
	"net/http"
)

// Server datastructure encapsolating data for connections
// and targets
type Server struct {
	addr    string
	clients map[string]clientConn
	targets map[string]map[string]int
}

type clientConn struct {
	name    string
	addr    string
	actions chan action
}

// NewServer creates a new instance of a server
func NewServer(addr string) *Server {
	return &Server{
		addr:    addr,
		clients: make(map[string]clientConn),
		targets: map[string]map[string]int{},
	}
}

// Start Listening on http for incoming client connections
func (s *Server) Start() error {
	http.HandleFunc("/join", s.join)
	http.HandleFunc("/request", s.request)
	http.HandleFunc("/report", s.report)
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
	s.targets[requestID] = map[string]int{}
	for _, client := range s.clients {
		client.actions <- action{
			RequestID: requestID,
			Type:      ping,
			Target:    target,
		}
	}
}

func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	var res response
	err := json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		http.Error(w, "Invalid response", http.StatusBadRequest)
		return
	}
	s.targets[res.RequestID][res.ClientName] = res.Ping
	log.Printf("Report received %v\n", s.targets)
}
