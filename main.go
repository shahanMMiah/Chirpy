package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func RedinisHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(200)

	stat, _ := resp.Write([]byte("OK"))

	fmt.Printf("%v - %v", resp.Header(), stat)
}

type ApiConfig struct {
	fileserverHits atomic.Int32
}

func (a *ApiConfig) MiddlewareIncHits(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		a.fileserverHits.Add(1)
		fmt.Printf("incrmenting Hit to : %v", a.fileserverHits.Load())
		handler.ServeHTTP(resp, req)
	})

}

func (a *ApiConfig) MiddlewareReqCheckHandle() func(http.ResponseWriter, *http.Request) {

	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf("Hits: %v ", a.fileserverHits.Load())))
	}

}

func (a *ApiConfig) MiddlewareReqResetHandle() func(http.ResponseWriter, *http.Request) {

	return func(resp http.ResponseWriter, req *http.Request) {
		a.fileserverHits.Store(0)
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf("Server hits reset to: %v ", a.fileserverHits.Load())))
	}

}

func main() {

	mux := http.NewServeMux()

	a := ApiConfig{}
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	handler := http.FileServer(http.Dir("."))

	mux.HandleFunc("GET /healthz", RedinisHandler)
	mux.HandleFunc("POST /reset", a.MiddlewareReqResetHandle())
	mux.HandleFunc("GET /metrics", a.MiddlewareReqCheckHandle())

	mux.Handle("/app/", a.MiddlewareIncHits(http.StripPrefix("/app", handler)))

	server.ListenAndServe()

}
