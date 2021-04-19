package main

import (
	"log"
	"net/http"
	"net/http/httptrace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
)

type index struct {
	dependantService string
	httpClient       http.Client
	serviceName      string
}

func (i index) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("received request to index")

	ctx := httptrace.WithClientTrace(req.Context(), otelhttptrace.NewClientTrace(req.Context()))
	backendRequest, _ := http.NewRequestWithContext(ctx, "GET", i.dependantService, nil)

	_, err := i.httpClient.Do(backendRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte("Hello from " + i.serviceName))
}
