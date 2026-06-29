package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"review-api/internal/repository"
	"review-api/internal/service"

	"review-api/internal/applefeed"

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
	reviewService := service.NewService(feed, repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /reviews", func(w http.ResponseWriter, r *http.Request) {
		reviews, err := reviewService.GetReviews(r.Context())
		if err != nil {
			log.Println("failed to fetch reviews:", err)
			http.Error(w, "failed to fetch reviews", http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(reviews)
		if err != nil {
			http.Error(w, "failed to encode reviews", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
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
