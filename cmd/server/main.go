package main

import (
	"log"

	"github.com/lehigh-university-libraries/mountain-hawk/internal/config"
	"github.com/lehigh-university-libraries/mountain-hawk/internal/server"
)

func main() {
	cfg := config.MustLoad()

	srv := server.New(cfg)

	log.Printf("Starting server on port %s", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
