package applefeed

// AppleFeed represents the JSON structure of the Apple App Store Review RSS feed.
// It has been generated automatically from the JSON content.
type AppleFeed struct {
	Feed AppleFeedData `json:"feed"`
}

type AppleFeedData struct {
	Author  AppleAuthor  `json:"author"`
	Entry   []AppleEntry `json:"entry"`
	Updated AppleLabel   `json:"updated"`
	Rights  AppleLabel   `json:"rights"`
	Title   AppleLabel   `json:"title"`
	Icon    AppleLabel   `json:"icon"`
	Link    []AppleLink  `json:"link"`
	ID      AppleLabel   `json:"id"`
}

type AppleAuthor struct {
	Name  AppleLabel `json:"name"`
	URI   AppleLabel `json:"uri"`
	Label string     `json:"label,omitempty"`
}

type AppleEntry struct {
	Author        AppleAuthor        `json:"author"`
	Updated       AppleLabel         `json:"updated"`
	ImRating      AppleLabel         `json:"im:rating"`
	ImVersion     AppleLabel         `json:"im:version"`
	ID            AppleLabel         `json:"id"`
	Title         AppleLabel         `json:"title"`
	Content       AppleContent       `json:"content"`
	Link          AppleLink          `json:"link"`
	ImVoteSum     AppleLabel         `json:"im:voteSum"`
	ImContentType AppleImContentType `json:"im:contentType"`
	ImVoteCount   AppleLabel         `json:"im:voteCount"`
}

type AppleLabel struct {
	Label string `json:"label"`
}

type AppleContent struct {
	Label      string                 `json:"label"`
	Attributes AppleContentAttributes `json:"attributes"`
}

type AppleContentAttributes struct {
	Type string `json:"type"`
}

type AppleLink struct {
	Attributes AppleLinkAttributes `json:"attributes"`
}

type AppleLinkAttributes struct {
	Rel  string `json:"rel,omitempty"`
	Type string `json:"type,omitempty"`
	Href string `json:"href,omitempty"`
}

type AppleImContentType struct {
	Attributes AppleImContentTypeAttributes `json:"attributes"`
}

type AppleImContentTypeAttributes struct {
	Term  string `json:"term"`
	Label string `json:"label"`
}
