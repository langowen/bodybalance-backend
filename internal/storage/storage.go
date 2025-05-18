package storage

// Video represents video item response structure

type Video struct {
	ID          float64 `json:"id"`
	URL         string  `json:"url"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
}
