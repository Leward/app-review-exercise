package repository

import (
	"review-api/internal/domain"
	"time"

	"gorm.io/gorm"
)

type AppleReview struct {
	SourceID string    `gorm:"primaryKey;column:source_id"`
	Title    string    `gorm:"column:title"`
	Author   string    `gorm:"column:author"`
	Content  string    `gorm:"column:content"`
	Rating   int       `gorm:"column:rating"`
	Date     time.Time `gorm:"column:date;index:idx_date"`
}

func newAppleReviewFromDomain(review domain.Review) AppleReview {
	return AppleReview{
		SourceID: review.SourceID,
		Title:    review.Title,
		Author:   review.Author,
		Content:  review.Content,
		Rating:   review.Rating,
		Date:     review.Date,
	}
}

func newAppleReviewsFromDomain(reviews []domain.Review) []AppleReview {
	var appleReviews []AppleReview
	for _, review := range reviews {
		appleReviews = append(appleReviews, newAppleReviewFromDomain(review))
	}
	return appleReviews
}

func (a AppleReview) TableName() string {
	return "reviews_apple"
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&AppleReview{})
}
