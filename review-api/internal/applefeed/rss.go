package applefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Source identifies the originating system for reviews fetched from this feed.
// It is used as the `source` part of the reviews composite primary key.
const Source = "apple"

// Review is the parsed representation of a single review from the Apple iTunes
// RSS feed. It is converted to a store.Review by the sync flow.
type Review struct {
	SourceID string
	Title    string
	Author   string
	Content  string
	Rating   int
	Date     time.Time
}

type Feed struct {
	id       string
	client   http.Client
	baseURL  string
	maxDepth int
}

func NewFeed(id string) *Feed {
	return &Feed{
		id:       id,
		client:   http.Client{},
		maxDepth: 10,
	}
}

// Fetch pulls the latest reviews from the feed. It satisfies the
// reviewsync.FeedFetcher interface.
func (f *Feed) Fetch(ctx context.Context, after *time.Time) ([]Review, error) {
	return f.collectReviews(after)
}

func (f *Feed) collectReviews(after *time.Time) ([]Review, error) {
	var allReviews []Review

	// Fetch first page to determine total pages
	firstPageURL := f.url(1)
	resp, err := f.client.Get(firstPageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get first page: %w", err)
	}
	pageReviews, totalPages, err := f.parseJSON(resp, after)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	allReviews = append(allReviews, pageReviews...)

	// Limit total pages to maxDepth
	if totalPages > f.maxDepth {
		totalPages = f.maxDepth
	}

	// Fetch remaining pages
	for page := 2; page <= totalPages; page++ {
		url := f.url(page)
		resp, err := f.client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to get page %d: %w", page, err)
		}
		pageReviews, _, err := f.parseJSON(resp, after)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		allReviews = append(allReviews, pageReviews...)
	}

	return allReviews, nil
}

func (f *Feed) parseJSON(resp *http.Response, after *time.Time) ([]Review, int, error) {
	var feed AppleFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, 0, fmt.Errorf("failed to decode JSON reviews: %w", err)
	}

	reviews := make([]Review, 0, len(feed.Feed.Entry))
	for _, entry := range feed.Feed.Entry {
		updated, err := time.Parse(time.RFC3339, entry.Updated.Label)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse updated date: %w", err)
		}
		if after != nil && !updated.After(*after) {
			continue
		}
		rating, err := strconv.Atoi(entry.ImRating.Label)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse rating %q: %w", entry.ImRating.Label, err)
		}

		reviews = append(reviews, Review{
			SourceID: entry.ID.Label,
			Title:    entry.Title.Label,
			Author:   entry.Author.Name.Label,
			Content:  entry.Content.Label,
			Rating:   rating,
			Date:     updated,
		})
	}

	return reviews, detectLastPage(feed), nil
}

func detectLastPage(feed AppleFeed) int {
	for _, link := range feed.Feed.Link {
		if link.Attributes.Rel == "last" {
			return extractPageNumber(link.Attributes.Href)
		}
	}
	// Fallback: check if "next" link exists
	for _, link := range feed.Feed.Link {
		if link.Attributes.Rel == "next" {
			return extractPageNumber(link.Attributes.Href)
		}
	}
	return 1 // default in case no last page link is found
}

func extractPageNumber(url string) int {
	page := 1
	// Example href: https://itunes.apple.com/us/rss/customerreviews/page=10/id=595068606/sortby=mostrecent/xml?urlDesc=/customerreviews/id=595068606/sortBy=mostRecent/page=1/json
	// Extract page number from /page=N/
	// We look for "page=" and then the number until the next "/"
	start := strings.Index(url, "page=")
	if start != -1 {
		start += 5
		end := strings.Index(url[start:], "/")
		if end != -1 {
			pageStr := url[start : start+end]
			fmt.Sscanf(pageStr, "%d", &page)
		}
	}
	return page
}

func (f *Feed) url(page int) string {
	base := f.baseURL
	if base == "" {
		base = "https://itunes.apple.com"
	}
	return fmt.Sprintf("%s/us/rss/customerreviews/id=%s/sortBy=mostRecent/page=%d/json", base, f.id, page)
}
