package api

import (
	"bytes"
	"log"
	"net/http"
)

func forwardRequestToDestination(r *http.Request, bodyBytes []byte, destination string, payload *dinkRequestPayload) {
	log.Printf("[API:forward] Sending request %q to %q\n", payload.Type, destination)

	outReq, _ := http.NewRequest(r.Method, destination, bytes.NewReader(bodyBytes))
	outReq.Header = r.Header.Clone()
	http.DefaultClient.Do(outReq)
}
