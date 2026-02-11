package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"autotrade/backend-go/internal/config"
	"autotrade/backend-go/internal/db"
	httpserver "autotrade/backend-go/internal/http"
	"autotrade/backend-go/internal/repo"
	"autotrade/backend-go/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(context.Background(), pool); err != nil {
		log.Fatalf("db migrate error: %v", err)
	}

	r := repo.New(pool)
	svc := service.New(r)
	h := httpserver.NewHandler(svc)

	go httpserver.ServeWS(":" + cfg.WsPort)

	addr := ":" + cfg.GoPort
	log.Printf("api listening on %s", addr)
	if err := http.ListenAndServe(addr, h.Router()); err != nil {
		log.Fatal(err)
	}
}
