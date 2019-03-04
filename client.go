package locust

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// Client data structure
type Client struct {
	name     string
	hostAddr string
	actions  chan action
}

// NewClient creates a new instance of a client connected
//  to host machine.
func NewClient(name, host string) *Client {
	return &Client{
		name:     name,
		hostAddr: host,
		actions:  make(chan action),
	}
}

// Start connects to host address via http and starts
// listening for actions from the response body
func (c *Client) Start() error {
	req, err := http.NewRequest("GET", c.hostAddr+"/join", nil)
	if err != nil {
		return err
	}
	req.Header.Add("client-name", c.name)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	go c.handleActions()
	return c.readActions(resp.Body)
}

func (c *Client) readActions(r io.Reader) error {
	dec := json.NewDecoder(r)
	for dec.More() {
		var a action
		err := dec.Decode(&a)
		if err != nil {
			return err
		}
		log.Printf("New action received %v\n", a)
		c.actions <- a
	}
	return nil
}

func (c *Client) handleActions() {
	for action := range c.actions {
		p, err := c.ping(action.Target)
		if err != nil {
			continue
		}
		c.reportPing(action, p)
	}
}

func (c *Client) ping(target string) (int, error) {
	start := time.Now()
	resp, err := http.Get(target)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	elapsed := time.Now().Sub(start)
	return int(elapsed), nil
}

func (c *Client) reportPing(a action, ping int) {
	var buff bytes.Buffer
	enc := json.NewEncoder(&buff)
	data := response{
		RequestID:  a.RequestID,
		ClientName: c.name,
		Ping:       ping,
	}
	enc.Encode(&data)
	resp, err := http.Post(c.hostAddr+"/report", "application/json", &buff)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
}
