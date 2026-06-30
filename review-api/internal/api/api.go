package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"review-api/internal/domain"
	"review-api/internal/service"
)

const (
	SSEEventData         = "data"
	SSEEventRefreshError = "refresh_error"
)

type ReviewsService interface {
	GetAppleReviews(context.Context) ([]domain.Review, error)
	GetAppleReviewsAsync(context.Context) <-chan service.AsyncRefreshResult
}

func NewHandler(reviewService ReviewsService) http.Handler {
	mux := http.NewServeMux()
	registerRoutes(mux, reviewService)
	return mux
}

func registerRoutes(mux *http.ServeMux, reviewService ReviewsService) {
	mux.HandleFunc("GET /reviews", func(w http.ResponseWriter, r *http.Request) {
		handleReviews(w, r, reviewService)
	})

	mux.HandleFunc("GET /reviews-async", func(w http.ResponseWriter, r *http.Request) {
		handleReviewsAsync(w, r, reviewService)
	})
}

func handleReviews(w http.ResponseWriter, r *http.Request, reviewService ReviewsService) {
	reviews, err := reviewService.GetAppleReviews(r.Context())
	if err != nil {
		log.Println("failed to fetch reviews:", err)
		http.Error(w, "failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, reviews); err != nil {
		http.Error(w, "failed to encode reviews", http.StatusInternalServerError)
	}
}

func handleReviewsAsync(w http.ResponseWriter, r *http.Request, reviewService ReviewsService) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	refreshResultCh := reviewService.GetAppleReviewsAsync(ctx)

	for result := range refreshResultCh {
		if result.Err != nil {
			if errors.Is(result.Err, context.Canceled) {
				return
			}

			if err := writeSSEEvent(w, SSEEventRefreshError, map[string]string{
				"error": result.Err.Error(),
			}); err != nil {
				log.Println("failed to write refresh_error event:", err)
			}
			flusher.Flush()
			return
		}

		if err := writeSSEEvent(w, SSEEventData, map[string]any{
			"reviews": result.Reviews,
		}); err != nil {
			log.Println("failed to write data event:", err)
			return
		}
		flusher.Flush()
	}
}

func writeJSON(w http.ResponseWriter, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

func writeSSEEvent(w io.Writer, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return fmt.Errorf("write event line: %w", err)
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("write data line: %w", err)
	}
	return nil
}
