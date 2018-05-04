package youtube

import (
	"time"
)

// Video holds basic info of a video
type Video struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Duration    time.Duration `json:"duration"`
	ViewCount   int           `json:"view_count"`
	PublishDate string        `json:"publish_date"`
	Thumbnails  Thumbnails    `json:"thumbnail"`
}

// Thumbnails holds info of several thumbnail images with different sizes
type Thumbnails struct {
	Default  ThumbnailDetail `json:"default"`
	Medium   ThumbnailDetail `json:"medium"`
	High     ThumbnailDetail `json:"high"`
	Standard ThumbnailDetail `json:"standard"`
	Maxres   ThumbnailDetail `json:"maxres"`
}

// ThumbnailDetail holds url and size of a thumbnail image
type ThumbnailDetail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
