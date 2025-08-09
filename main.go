package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

const (
	ollamaAddr = "http://localhost"
)

func main() {
	target, _ := url.Parse(ollamaAddr)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
	}

	Backend = NewModelBackend()
	ctx := context.Background()
	model := os.Getenv("MODEL")
	err := Backend.RunModel(ctx, model)
	if err != nil {
		panic(err)
	}
	select {
	case <-Backend.StartupDone.Done():

	case <-ctx.Done():
		panic(errors.New("ctx is calceled"))
	}
	log.Printf("model started %s", model)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/attest" {
			AttestHandler(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	log.Printf("Go proxy started on %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("http server quite: %v", err)
	}
}
