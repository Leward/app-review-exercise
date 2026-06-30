package service

import (
	"context"
	"fmt"
	"review-api/internal/domain"
)

type Repository interface {
	List(ctx context.Context) ([]domain.Review, error)
}

type Syncer interface {
	SyncAppleReviews(context.Context) ([]domain.Review, error)
}

type Service struct {
	sync Syncer
	repo Repository
}
type AsyncRefreshResult struct {
	Reviews []domain.Review
	Err     error
}

func NewService(sync Syncer, repo Repository) *Service {
	return &Service{
		sync: sync,
		repo: repo,
	}
}

// GetAppleReviews returns the latest reviews from the feed and the local database.
func (s *Service) GetAppleReviews(ctx context.Context) ([]domain.Review, error) {
	reviews, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list reviews: %w", err)
	}

	newReviews, err := s.sync.SyncAppleReviews(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews from feed: %w", err)
	}

	return append(newReviews, reviews...), nil
}

// GetAppleReviewsAsync returns reviews in two steps asynchronously:
// first the DB snapshot, then only newly fetched reviews.
func (s *Service) GetAppleReviewsAsync(ctx context.Context) <-chan AsyncRefreshResult {
	resultCh := make(chan AsyncRefreshResult, 2)

	go func() {
		defer close(resultCh)

		// 1. Return DB reviews immediately (fast)
		reviews, err := s.repo.List(ctx)
		if err != nil {
			resultCh <- AsyncRefreshResult{Err: fmt.Errorf("failed to list reviews: %w", err)}
			return
		}
		resultCh <- AsyncRefreshResult{Reviews: reviews}

		// 2. Sync from Apple (slow), emit only newly fetched reviews
		newReviews, err := s.sync.SyncAppleReviews(ctx)
		if err != nil {
			resultCh <- AsyncRefreshResult{Err: err}
			return
		}
		resultCh <- AsyncRefreshResult{Reviews: newReviews}
	}()

	return resultCh
}
