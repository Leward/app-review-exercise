package service

import (
	"context"
	"fmt"
	"review-api/internal/domain"
	"time"
)

type Feed interface {
	Fetch(context.Context, *time.Time) ([]domain.Review, error)
}

type Repository interface {
	Persist(context.Context, []domain.Review) error
}

type Service struct {
	feed Feed
	repo Repository
}

func NewService(feed Feed, repo Repository) *Service {
	return &Service{
		feed: feed,
		repo: repo,
	}
}

func (s *Service) GetReviews(ctx context.Context) ([]domain.Review, error) {
	//since := time.Now().Add(-48 * time.Hour)
	feedReviews, err := s.feed.Fetch(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews from feed: %w", err)
	}

	if err := s.repo.Persist(ctx, feedReviews); err != nil {
		return nil, fmt.Errorf("failed to persist reviews: %w", err)
	}

	return feedReviews, nil
}
