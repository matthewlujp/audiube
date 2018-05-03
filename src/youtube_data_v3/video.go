package youtube

import (
	"time"
)

type Video struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Duration    time.Duration `json:"duration"`
	ViewCount   int           `json:"view_count"`
	PublishDate string        `json:"publish_date"`
	Thumbnails  Thumbnails    `json:"thumbnail"`
}

type Thumbnails struct {
	Default  ThumbnailDetail `json:"default"`
	Medium   ThumbnailDetail `json:"medium"`
	High     ThumbnailDetail `json:"high"`
	Standard ThumbnailDetail `json:"standard"`
	Maxres   ThumbnailDetail `json:"maxres"`
}

type ThumbnailDetail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
