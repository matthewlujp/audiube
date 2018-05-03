package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// durationRegex = regexp.MustCompile(`PT(?P<hour>[0-9]+H)?(?P<minute>[0-9]+M)?(?P<second>[0-9]+S)`)
	durationRegex = regexp.MustCompile(`PT([0-9HMS]+)`)
)

func parseSearchResult(r io.ReadCloser) ([]string, error) {
	// define a struct which is compatible with the search response json
	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed parsing search result, %s", err)
	}

	ids := make([]string, 0, len(result.Items))
	for _, i := range result.Items {
		ids = append(ids, i.ID.VideoID)
	}
	return ids, nil
}

func parseVideosDetails(r io.ReadCloser) ([]Video, error) {
	// define a struct which is compatible with the video details response json
	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				PublishedAt string `json:"publishedAt"`
				Title       string `json:"title"`
				Thumbnails  struct {
					Default struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"default"`
					Medium struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"medium"`
					High struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"high"`
					Standard struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"standard"`
					Maxres struct {
						URL    string `json:"url"`
						Width  int    `json:"width"`
						Height int    `json:"height"`
					} `json:"maxres"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	// decode the reader into the structure
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return nil, err
	}

	// set retrieved info into a []Video
	videos := make([]Video, 0, len(result.Items))
	for _, item := range result.Items {
		viewCount, _ := strconv.Atoi(item.Statistics.ViewCount)
		duration := parseDuration(item.ContentDetails.Duration)
		v := Video{
			ID:          item.ID,
			Title:       item.Snippet.Title,
			Duration:    duration,
			ViewCount:   viewCount,
			PublishDate: strings.Split(item.Snippet.PublishedAt, "T")[0],
			Thumbnails: Thumbnails{
				Default: ThumbnailDetail{
					URL:    item.Snippet.Thumbnails.Default.URL,
					Width:  item.Snippet.Thumbnails.Default.Width,
					Height: item.Snippet.Thumbnails.Default.Height,
				},
				Medium: ThumbnailDetail{
					URL:    item.Snippet.Thumbnails.Medium.URL,
					Width:  item.Snippet.Thumbnails.Medium.Width,
					Height: item.Snippet.Thumbnails.Medium.Height,
				},
				High: ThumbnailDetail{
					URL:    item.Snippet.Thumbnails.High.URL,
					Width:  item.Snippet.Thumbnails.High.Width,
					Height: item.Snippet.Thumbnails.High.Height,
				},
				Standard: ThumbnailDetail{
					URL:    item.Snippet.Thumbnails.Standard.URL,
					Width:  item.Snippet.Thumbnails.Standard.Width,
					Height: item.Snippet.Thumbnails.Standard.Height,
				},
				Maxres: ThumbnailDetail{
					URL:    item.Snippet.Thumbnails.Maxres.URL,
					Width:  item.Snippet.Thumbnails.Maxres.Width,
					Height: item.Snippet.Thumbnails.Maxres.Height,
				},
			},
		}
		videos = append(videos, v)
	}

	return videos, nil
}

// parseDuration converts PT1H4M9S -> time.Hour*1 + time.Minute*4 + time.Second*9
func parseDuration(dur string) time.Duration {
	matched := durationRegex.FindStringSubmatch(dur)
	if matched == nil {
		return time.Duration(0)
	}
	// matched[1] is something like "1H12M42S"

	// convert a formatted duration "1H12M42S" into time.Duration
	duration, _ := time.ParseDuration(strings.ToLower(matched[1])) // ParseDuration only deals lowercase units
	return duration
}
