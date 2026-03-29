package source

import "time"

type ManualSource struct {
	URL string
}

func NewManualSource(url string) *ManualSource {
	return &ManualSource{URL: url}
}

func (m *ManualSource) Name() string { return "manual" }

func (m *ManualSource) Fetch() ([]ReadingItem, error) {
	return []ReadingItem{
		{
			URL:       m.URL,
			Source:    "manual",
			DateAdded: time.Now(),
			IsUnread:  true,
		},
	}, nil
}
