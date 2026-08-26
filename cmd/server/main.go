package main

import (
	"flag"
	"log"
	"net/http"

	"bannerfingerprint/internal/engine"
	"bannerfingerprint/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	rules := flag.String("rules", "", "optional path to an external fingerprints.yaml overriding the embedded set")
	flag.Parse()

	var e *engine.Engine
	var err error
	if *rules != "" {
		e, err = engine.NewFromFile(*rules)
	} else {
		e, err = engine.New()
	}
	if err != nil {
		log.Fatalf("init engine: %v", err)
	}

	log.Printf("banner fingerprint server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, server.NewHandler(e)); err != nil {
		log.Fatalf("server: %v", err)
	}
}
