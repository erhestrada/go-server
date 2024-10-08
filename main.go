package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
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
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf("<html><body>Page visited %d times</body></html>", config.fileserverHits)))
	//fmt.Fprint(w, config.fileserverHits)
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

func validateJson(w http.ResponseWriter, req *http.Request) {
	validJsonRequest := decodeJson(w, req)
	if validJsonRequest == 200 {
		encodeJsonValidRequest(w, true)
	} else if validJsonRequest == 400 {
		encodeJsonInvalidRequest(w)
	}
}

func decodeJson(w http.ResponseWriter, req *http.Request) int {
	type parameters struct {
		Body string `json:"body"`
	}

	maxCharacters := 140

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	//fmt.Println(params)
	//fmt.Println(params.Body)
	//fmt.Println(len(params.Body))

	if err != nil {
		//log.Printf("Error decoding parameters: %s", err)
		log.Printf("something went wrong :/")
		w.WriteHeader(500)
		return 500
	}

	if len(params.Body) > maxCharacters {
		log.Printf("too many characters >:(")
		w.WriteHeader(400)
		return 400
	}

	fmt.Println(params)
	return 200
}

func encodeJsonValidRequest(w http.ResponseWriter, valid bool) {
	type returnVals struct {
		CreatedAt time.Time `json:"created_at"`
		ID        int       `json:"id"`
		Valid     bool      `json:"valid"`
	}
	respBody := returnVals{
		CreatedAt: time.Now(),
		ID:        123,
		Valid:     valid,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func encodeJsonInvalidRequest(w http.ResponseWriter) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: "message too long :/",
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func main() {
	cfg := &apiConfig{fileserverHits: atomic.Int32{}}

	serveMux := http.NewServeMux()
	//serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.Handle("/app/assets/catpfp.jpg", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("GET /api/healthz", checkReadiness)
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /admin/metrics", cfg.checkHits)
	serveMux.HandleFunc("POST /admin/reset", cfg.resetHits)
	serveMux.HandleFunc("POST /api/validate_json", validateJson)

	server := &http.Server{Addr: ":8080",
		Handler: serveMux}
	server.ListenAndServe()
}
