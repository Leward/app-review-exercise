package repository

import (
	"context"
	"errors"
	"review-api/internal/domain"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppleReviewRepo struct {
	db *gorm.DB
}

func NewAppleReview(db *gorm.DB) *AppleReviewRepo {
	return &AppleReviewRepo{
		db: db,
	}
}

func (r *AppleReviewRepo) Persist(context context.Context, reviews []domain.Review) error {
	return r.db.WithContext(context).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(appleReviewsFromDomain(reviews)).Error
}

func (r *AppleReviewRepo) LatestReviewDate(ctx context.Context) (*time.Time, error) {
	var latestReviewDate *time.Time
	err := r.db.WithContext(ctx).Order("date desc").First(&latestReviewDate).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return latestReviewDate, nil
}

func (r *AppleReviewRepo) List(ctx context.Context) ([]domain.Review, error) {
	var reviews []AppleReview
	err := r.db.WithContext(ctx).Find(&reviews).Error
	if err != nil {
		return nil, err
	}
	return appleReviewsToDomain(reviews), nil
}
