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

func NewService(sync Syncer, repo Repository) *Service {
	return &Service{
		sync: sync,
		repo: repo,
	}
}

func (s *Service) GetAppleReviews(ctx context.Context) ([]domain.Review, error) {

	// Fetch existing reviews from DB
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
