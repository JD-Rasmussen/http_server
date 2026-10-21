package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

type Chirp struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) getFileserverHits() int32 {
	return cfg.fileserverHits.Load()
}

func (cfg *apiConfig) resetFileserverHits() {
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}
func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func main() {
	apiCfg := &apiConfig{}

	mux := http.ServeMux{}

	server := http.Server{
		Addr:    ":8080",
		Handler: &mux,
	}

	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
	/*
		mux.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("Fileserver Hits: %d", apiCfg.getFileserverHits())))
		})
	*/
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", apiCfg.getFileserverHits())))
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		apiCfg.resetFileserverHits()
		w.Write([]byte("Fileserver Hits reset"))
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		chirp := Chirp{}
		err := decoder.Decode(&chirp)
		if err != nil {
			log.Printf("Error decoding chirp: %s", err)
			respondWithError(w, http.StatusInternalServerError, "Something went wrong")
			return
		}
		if len(chirp.Body) > 140 {
			respondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"valid": true})
	})

	log.Fatal(server.ListenAndServe())

}
