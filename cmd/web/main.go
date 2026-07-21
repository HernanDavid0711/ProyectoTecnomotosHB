package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tecnomotos/internal/config"
	"tecnomotos/internal/database"
	"tecnomotos/internal/httpserver"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error cargando configuración: %v", err)
	}

	db, err := database.NewPostgresPool(context.Background(), cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("error conectando a postgres: %v", err)
	}
	defer db.Close()

	log.Println("base de datos conectada correctamente")

	server := httpserver.New(cfg, db)

	go func() {
		log.Printf("servidor escuchando en puerto %s", cfg.HTTPPort)
		if err := server.Start(); err != nil {
			log.Fatalf("error iniciando servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("apagando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("error apagando servidor: %v", err)
	}

	log.Println("servidor detenido correctamente")
}