package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func checkReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (config *apiConfig) checkHits(w http.ResponseWriter, req *http.Request) {
	//w.Write([]byte("OK"))
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, config.fileserverHits)
	//w.Write([]byte(config.fileserverHits))
}

func (config *apiConfig) resetHits(w http.ResponseWriter, req *http.Request) {
	config.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	cfg := &apiConfig{fileserverHits: atomic.Int32{}}

	serveMux := http.NewServeMux()
	//serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.Handle("/app/assets/catpfp.jpg", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("GET /healthz", checkReadiness)
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /metrics", cfg.checkHits)
	serveMux.HandleFunc("POST /reset", cfg.resetHits)

	server := &http.Server{Addr: ":8080",
		Handler: serveMux}
	server.ListenAndServe()
}
