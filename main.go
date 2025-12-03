package main

import "net/http"

func main() {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/healthz", readinessEndpoint)
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}

	serveMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("/home/josh/Documents/repos/github.com/joshckidd/chirpy"))))
	server.ListenAndServe()
}
