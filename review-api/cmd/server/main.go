package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"

	"review-api/internal/applefeed"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /reviews", func(w http.ResponseWriter, r *http.Request) {
		feed := applefeed.NewFeed("595068606")

		reviews, err := feed.Fetch(r.Context(), nil)
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
		w.Write(data)
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
	log.Println("Server listening on :8080")

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background()) // we may want to enforce a max shutdown delay
	}()

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start server:", err)
	}
}
