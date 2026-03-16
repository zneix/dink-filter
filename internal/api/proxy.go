package api

import (
	"log"
	"net/http"
)

func forwardRequestToDestination(r *http.Request, destination string, dr *dinkRequest) {
	log.Printf("[API:forward] Sending request %q to %q\n", dr.payload.Type, destination)

	outReq, _ := http.NewRequest(r.Method, destination, r.Body)
	outReq.Header = r.Header.Clone()
	http.DefaultClient.Do(outReq)
}
