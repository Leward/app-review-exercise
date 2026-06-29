package sync

import (
	"context"
	"fmt"
	"review-api/internal/domain"
	"time"
)

type Repository interface {
	Persist(context.Context, []domain.Review) error
	LatestReviewDate(ctx context.Context) (*time.Time, error)
}

type Feed interface {
	Fetch(context.Context, *time.Time) ([]domain.Review, error)
}

type Sync struct {
	repo Repository
	feed Feed
}

func New(repo Repository, feed Feed) Sync {
	return Sync{
		repo: repo,
		feed: feed,
	}
}

// SyncAppleReviews fetches new reviews from the Apple feed and persists them in the repository.
// It returns the new reviews that were fetched and persisted.
func (s Sync) SyncAppleReviews(ctx context.Context) ([]domain.Review, error) {
	feed := s.feed
	since, err := s.repo.LatestReviewDate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest review date: %w", err)
	}

	reviews, err := feed.Fetch(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews: %w", err)
	}

	if err := s.repo.Persist(ctx, reviews); err != nil {
		return nil, fmt.Errorf("failed to persist reviews: %w", err)
	}
	return reviews, nil
}
