package main

import (
	"net/http"
)

func checkReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

func main() {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/assets/catpfp.jpg", http.FileServer(http.Dir(".")))
	serveMux.HandleFunc("/healthz", checkReadiness)
	server := &http.Server{Addr: ":8080",
		Handler: serveMux}
	server.ListenAndServe()
}
