package repository

import (
	"context"
	"review-api/internal/domain"

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
		Create(newAppleReviewsFromDomain(reviews)).Error
}
