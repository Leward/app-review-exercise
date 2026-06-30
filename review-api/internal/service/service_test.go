package service

import (
	"context"
	"errors"
	"review-api/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	listReviews []domain.Review
	listErr     error
	listCalls   int
}

func (f *fakeRepo) List(context.Context) ([]domain.Review, error) {
	f.listCalls++
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listReviews, nil
}

type fakeSync struct {
	reviews []domain.Review
	err     error
	calls   int
}

func (f *fakeSync) SyncAppleReviews(context.Context) ([]domain.Review, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.reviews, nil
}

func TestServiceGetAppleReviewsAsync(t *testing.T) {
	now := time.Date(2026, 6, 4, 9, 0, 0, 0, time.UTC)
	tests := []struct {
		name                  string
		repo                  fakeRepo
		sync                  fakeSync
		expectedReviewBatches [][]domain.Review
		expectedErrs          []string
		expectedListCalls     int
		expectedSyncCalls     int
	}{
		{
			name: "emits db snapshot then newly fetched reviews",
			repo: fakeRepo{
				listReviews: []domain.Review{
					{SourceID: "db-1", Date: now},
				},
			},
			sync: fakeSync{
				reviews: []domain.Review{
					{SourceID: "new-1", Date: now.Add(time.Minute)},
					{SourceID: "new-2", Date: now.Add(2 * time.Minute)},
				},
			},
			expectedReviewBatches: [][]domain.Review{
				{
					{SourceID: "db-1", Date: now},
				},
				{
					{SourceID: "new-1", Date: now.Add(time.Minute)},
					{SourceID: "new-2", Date: now.Add(2 * time.Minute)},
				},
			},
			expectedErrs:      []string{"", ""},
			expectedListCalls: 1,
			expectedSyncCalls: 1,
		},
		{
			name: "returns list error and skips refresh",
			repo: fakeRepo{
				listErr: errors.New("db down"),
			},
			sync: fakeSync{
				reviews: []domain.Review{{SourceID: "new-1", Date: now}},
			},
			expectedReviewBatches: [][]domain.Review{
				nil,
			},
			expectedErrs: []string{
				"failed to list reviews: db down",
			},
			expectedListCalls: 1,
			expectedSyncCalls: 0,
		},
		{
			name: "returns refresh error after snapshot",
			repo: fakeRepo{
				listReviews: []domain.Review{{SourceID: "db-1", Date: now}},
			},
			sync: fakeSync{
				err: errors.New("feed timeout"),
			},
			expectedReviewBatches: [][]domain.Review{
				{{SourceID: "db-1", Date: now}},
				nil,
			},
			expectedErrs: []string{
				"",
				"feed timeout",
			},
			expectedListCalls: 1,
			expectedSyncCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			syncer := tt.sync
			svc := NewService(&syncer, &repo)

			results := make([]AsyncRefreshResult, 0, len(tt.expectedErrs))
			for result := range svc.GetAppleReviewsAsync(context.Background()) {
				results = append(results, result)
			}

			require.Len(t, results, len(tt.expectedErrs))
			for idx, expectedErr := range tt.expectedErrs {
				result := results[idx]
				if expectedErr != "" {
					require.Error(t, result.Err)
					assert.Equal(t, expectedErr, result.Err.Error())
					continue
				}

				require.NoError(t, result.Err)
				assert.Equal(t, tt.expectedReviewBatches[idx], result.Reviews)
			}

			assert.Equal(t, tt.expectedListCalls, repo.listCalls)
			assert.Equal(t, tt.expectedSyncCalls, syncer.calls)
		})
	}
}

func TestServiceGetAppleReviews(t *testing.T) {
	dbReviewDate := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	newReviewDate := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name              string
		repo              fakeRepo
		sync              fakeSync
		expectedSourceIDs []string
		expectedErr       string
	}{
		{
			name: "prepends freshly fetched reviews before existing db reviews",
			repo: fakeRepo{
				listReviews: []domain.Review{
					{SourceID: "db-1", Date: dbReviewDate},
					{SourceID: "db-2", Date: dbReviewDate.Add(-time.Hour)},
				},
			},
			sync: fakeSync{
				reviews: []domain.Review{
					{SourceID: "new-1", Date: newReviewDate},
				},
			},
			expectedSourceIDs: []string{"new-1", "db-1", "db-2"},
		},
		{
			name: "returns repository errors",
			repo: fakeRepo{
				listErr: errors.New("list failed"),
			},
			sync: fakeSync{
				reviews: []domain.Review{{SourceID: "new-1", Date: newReviewDate}},
			},
			expectedErr: "failed to list reviews: list failed",
		},
		{
			name: "returns sync errors",
			repo: fakeRepo{
				listReviews: []domain.Review{{SourceID: "db-1", Date: dbReviewDate}},
			},
			sync: fakeSync{
				err: errors.New("fetch failed"),
			},
			expectedErr: "failed to fetch reviews from feed: fetch failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			syncer := tt.sync
			svc := NewService(&syncer, &repo)

			reviews, err := svc.GetAppleReviews(context.Background())

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}

			require.NoError(t, err)
			require.Len(t, reviews, len(tt.expectedSourceIDs))
			for idx, sourceID := range tt.expectedSourceIDs {
				assert.Equal(t, sourceID, reviews[idx].SourceID)
			}
		})
	}
}
