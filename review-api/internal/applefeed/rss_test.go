package applefeed

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectReviews(t *testing.T) {
	// Setup mock server
	ms := NewMockServer()
	defer ms.Close()

	// Initialize Feed with mock server URL
	feed := NewFeed("595068606")
	feed.baseURL = ms.URL()

	// Call collectReviews
	reviews, err := feed.collectReviews(nil)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, 15, len(reviews))

	// Validate the first review (from page1.json)
	expectedDate, _ := time.Parse(time.RFC3339, "2026-06-08T20:04:18-07:00")
	assert.Equal(t, "Heudncomejbfbfjnx", reviews[0].Author)
	assert.Equal(t, "I absolutely love this app, it makes group dinners so easy to spilt. I recommend it to all my friends!", reviews[0].Content)
	assert.True(t, expectedDate.Equal(reviews[0].Date))
}

func TestCollectReviewsAfter(t *testing.T) {
	// Setup mock server
	ms := NewMockServer()
	defer ms.Close()

	// Initialize Feed with mock server URL
	feed := NewFeed("595068606")
	feed.baseURL = ms.URL()

	after, err := time.Parse(time.RFC3339, "2026-05-15T22:10:42-07:00")
	require.NoError(t, err)

	reviews, err := feed.collectReviews(&after)
	require.NoError(t, err)
	require.Len(t, reviews, 4)

	for _, review := range reviews {
		assert.True(t, review.Date.After(after))
	}

	expectedNewest, err := time.Parse(time.RFC3339, "2026-06-08T20:04:18-07:00")
	require.NoError(t, err)
	expectedOldest, err := time.Parse(time.RFC3339, "2026-05-16T09:38:36-07:00")
	require.NoError(t, err)
	assert.True(t, expectedNewest.Equal(reviews[0].Date))
	assert.True(t, expectedOldest.Equal(reviews[len(reviews)-1].Date))
}

func TestDetectLastPage(t *testing.T) {
	tests := []struct {
		name     string
		feed     AppleFeed
		expected int
	}{
		{
			name:     "no last page link",
			feed:     AppleFeed{},
			expected: 1,
		},
		{
			name: "10 pages",
			feed: AppleFeed{
				Feed: AppleFeedData{
					Link: []AppleLink{
						{
							Attributes: AppleLinkAttributes{
								Rel:  "last",
								Href: "https://itunes.apple.com/us/rss/customerreviews/page=10/id=595068606/sortby=mostrecent/xml",
							},
						},
					},
				},
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLastPage(tt.feed)
			assert.Equal(t, tt.expected, result)
		})
	}
}
