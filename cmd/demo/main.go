// Command demo serves the fleetctl demo console and the component sheet.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/markopolo123/baud-ui/demo"
)

func main() {
	addr := flag.String("addr", "localhost:8866", "listen address")
	flag.Parse()
	log.Printf("baud/ui demo: http://%s/ (sheet at /sheet)", *addr)
	log.Fatal(http.ListenAndServe(*addr, demo.NewMux()))
}
