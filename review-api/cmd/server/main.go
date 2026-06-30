package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"review-api/internal/api"
	"review-api/internal/applefeed"
	"review-api/internal/repository"
	"review-api/internal/service"
	"review-api/internal/sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Init gorm DB
	db, err := gorm.Open(sqlite.Open("reviews.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	err = repository.AutoMigrate(db)
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	// Init service and components
	repo := repository.NewAppleReview(db)
	feed := applefeed.NewFeed("595068606")
	syncer := sync.New(repo, feed)
	reviewService := service.NewService(syncer, repo)
	server := &http.Server{Addr: ":8080", Handler: api.NewHandler(reviewService)}
	log.Println("Server listening on :8080")

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background()) // we may want to enforce a max shutdown delay
	}()

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start server:", err)
	}
}
