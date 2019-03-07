package main

import (
	"flag"
	"log"

	"github.com/ccutch/locust"
)

func main() {
	isMaster := flag.Bool("master", false, "is master server")
	hostAddr := flag.String("hostAddr", "http://127.0.0.1:3000", "host address for client to connect to")
	clientName := flag.String("clientName", "test", "name of client")
	flag.Parse()

	var err error
	if *isMaster {
		server := locust.NewServer(":3000", true)
		err = server.Start()
	} else {
		client := locust.NewClient(*clientName, *hostAddr)
		err = client.Start()
	}

	log.Fatal("Error occurred, shutting down", err)
}
