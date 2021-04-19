package main

import (
	"net/http"
)

type index struct {
	dependantService string
	httpClient       http.Client
	serviceName      string
}

func (i index) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	backendRequest, _ := http.NewRequestWithContext(req.Context(), "GET", i.dependantService, nil)

	_, err := i.httpClient.Do(backendRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte("Hello from " + i.serviceName))
}
