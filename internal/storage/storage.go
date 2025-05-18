package storage

// Video represents video item response structure

type Video struct {
	ID          float64 `json:"id"`
	URL         string  `json:"url"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

// CategoryWithVideos представляет категорию с ID из БД
type CategoryWithVideos struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Videos []Video `json:"videos,omitempty"`
}
