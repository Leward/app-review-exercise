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

type Service struct {
	feed Feed
}

func NewService(feed Feed) *Service {
	return &Service{
		feed: feed,
	}
}

func (s *Service) GetReviews(ctx context.Context) ([]domain.Review, error) {
	//since := time.Now().Add(-48 * time.Hour)
	feedReviews, err := s.feed.Fetch(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews from feed: %w", err)
	}
	return feedReviews, nil
}
