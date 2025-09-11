package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/dnonakolesax/noted-auth/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	port := flag.Int("port", 8808, "port to run the server on")
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", *port)), // The url pointing to API definition
	))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
