// Command static-server serves an already-built release-candidate directory.
// It deliberately does not invoke Hugo, so checks exercise the production artifact.
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	directory := flag.String("dir", "public", "artifact directory to serve")
	address := flag.String("addr", "127.0.0.1:14140", "listen address")
	flag.Parse()

	server := &http.Server{
		Addr:              *address,
		Handler:           http.FileServer(http.Dir(*directory)),
		ReadHeaderTimeout: 5e9,
	}
	log.Fatal(server.ListenAndServe())
}
