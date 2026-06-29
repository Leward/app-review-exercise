package domain

import "time"

type Review struct {
	SourceID string
	Title    string
	Author   string
	Content  string
	Rating   int
	Date     time.Time
}
