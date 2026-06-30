package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"review-api/internal/domain"
	"review-api/internal/service"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeServerReviewsService struct {
	getReviews  []domain.Review
	getErr      error
	asyncStages []service.AsyncRefreshResult
	asyncDelay  time.Duration
}

func (f *fakeServerReviewsService) GetAppleReviews(context.Context) ([]domain.Review, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getReviews, nil
}

func (f *fakeServerReviewsService) GetAppleReviewsAsync(ctx context.Context) <-chan service.AsyncRefreshResult {
	resultCh := make(chan service.AsyncRefreshResult, len(f.asyncStages))

	go func() {
		defer close(resultCh)

		for _, stage := range f.asyncStages {
			if f.asyncDelay > 0 {
				select {
				case <-ctx.Done():
					resultCh <- service.AsyncRefreshResult{Err: ctx.Err()}
					return
				case <-time.After(f.asyncDelay):
				}
			}
			resultCh <- stage
		}
	}()

	return resultCh
}

func TestHandleReviews(t *testing.T) {
	now := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name                string
		service             fakeServerReviewsService
		expectedStatus      int
		expectedSourceIDs   []string
		expectedBodyContain string
	}{
		{
			name: "returns JSON reviews from service",
			service: fakeServerReviewsService{
				getReviews: []domain.Review{
					{SourceID: "new-1", Date: now},
					{SourceID: "db-1", Date: now.Add(-time.Minute)},
				},
			},
			expectedStatus:    http.StatusOK,
			expectedSourceIDs: []string{"new-1", "db-1"},
		},
		{
			name: "returns 500 when service fails",
			service: fakeServerReviewsService{
				getErr: errors.New("db down"),
			},
			expectedStatus:      http.StatusInternalServerError,
			expectedBodyContain: "failed to fetch reviews",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/reviews", nil)
			recorder := httptest.NewRecorder()
			handleReviews(recorder, req, &tt.service)

			require.Equal(t, tt.expectedStatus, recorder.Code)
			if tt.expectedStatus != http.StatusOK {
				assert.Contains(t, recorder.Body.String(), tt.expectedBodyContain)
				return
			}

			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			var reviews []domain.Review
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &reviews))
			assert.Equal(t, tt.expectedSourceIDs, sourceIDs(reviews))
		})
	}
}

func TestHandleReviewsAsyncEventOrdering(t *testing.T) {
	now := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		service        fakeServerReviewsService
		expectedEvents []string
	}{
		{
			name: "db snapshot before refreshed",
			service: fakeServerReviewsService{
				asyncStages: []service.AsyncRefreshResult{
					{
						Reviews: []domain.Review{{SourceID: "db-1", Date: now}},
					},
					{
						Reviews: []domain.Review{{SourceID: "new-1", Date: now.Add(time.Minute)}},
					},
				},
				asyncDelay: 5 * time.Millisecond,
			},
			expectedEvents: []string{SSEEventData, SSEEventData},
		},
		{
			name: "db snapshot followed by refresh error",
			service: fakeServerReviewsService{
				asyncStages: []service.AsyncRefreshResult{
					{
						Reviews: []domain.Review{{SourceID: "db-1", Date: now}},
					},
					{
						Err: errors.New("feed timeout"),
					},
				},
				asyncDelay: 5 * time.Millisecond,
			},
			expectedEvents: []string{SSEEventData, SSEEventRefreshError},
		},
		{
			name: "initial stream failure emits refresh error",
			service: fakeServerReviewsService{
				asyncStages: []service.AsyncRefreshResult{
					{
						Err: errors.New("db unavailable"),
					},
				},
			},
			expectedEvents: []string{SSEEventRefreshError},
		},
		{
			name: "context cancellation emits no stream event",
			service: fakeServerReviewsService{
				asyncStages: []service.AsyncRefreshResult{
					{
						Err: context.Canceled,
					},
				},
			},
			expectedEvents: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/reviews-async", nil)
			recorder := httptest.NewRecorder()
			handleReviewsAsync(recorder, req, &tt.service)

			require.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
			assert.Equal(t, tt.expectedEvents, sseEventNames(recorder.Body.String()))
		})
	}
}

func sseEventNames(payload string) []string {
	lines := strings.Split(payload, "\n")
	events := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "event: ") {
			events = append(events, strings.TrimPrefix(line, "event: "))
		}
	}
	return events
}

func sourceIDs(reviews []domain.Review) []string {
	ids := make([]string, 0, len(reviews))
	for _, review := range reviews {
		ids = append(ids, review.SourceID)
	}
	return ids
}
