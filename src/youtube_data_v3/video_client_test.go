package youtube

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock "github.com/matthewlujp/audiube/src/youtube_data_v3/mocks"
)

const (
	searchResultJSONPath   = "./mocks/data/search_result.json"
	searchDetailsJSONPath  = "./mocks/data/search_details.json"
	relatedResultJSONPath  = "./mocks/data/related_result.json"
	relatedDetailsJSONPath = "./mocks/data/related_details.json"
)

func TestSearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockClient(ctrl)

	apiKey := "foobar"
	req, errReq := http.NewRequest("GET", fmt.Sprintf(searchURL, "violet", 3, apiKey), nil)
	if errReq != nil {
		t.Fatal(errReq)
	}
	// request for video id search
	c.EXPECT().Do(req).Return(returnFileAsResponse(searchResultJSONPath))

	// request for video details
	req, errReq = http.NewRequest("GET", fmt.Sprintf(videoInfoURL, "UZxz9ot7y0Y,Ag4DR-L_TlM,nzdDUg5R_IQ", apiKey), nil)
	c.EXPECT().Do(req).Return(returnFileAsResponse(searchDetailsJSONPath))

	DefaultVideoClient.apiKey = apiKey
	DefaultVideoClient.client = c

	// check search result
	if videos, err := DefaultVideoClient.Search([]string{"violet"}, 3); err != nil {
		t.Errorf("search failed, %s", err)
	} else {
		// check only number of obtained videos and test one video info details
		if len(videos) != 3 {
			t.Errorf("number of videos expected %d, got %d", 3, len(videos))
		}
		for _, v := range videos {
			// check video info details for id = UZxz9ot7y0Y
			if v.ID == "UZxz9ot7y0Y" {
				if v.Title != "Alisson Shore - Violet Ft. JMakata, Colt" {
					t.Errorf("id=%s: title expected %s, got %s ", v.ID, "Alisson Shore - Violet Ft. JMakata, Colt", v.Title)
				}
				if v.Duration != time.Minute*4+time.Second*28 {
					t.Errorf("id=%s: duration expected %s, got %s ", v.ID, time.Minute*4+time.Second*28, v.Duration)
				}
				if v.ViewCount != 264829 {
					t.Errorf("id=%s: view count expected %d, got %d ", v.ID, 264328, v.ViewCount)
				}
				if v.PublishDate != "2018-02-16" {
					t.Errorf("id=%s: publish date expected %s, got %s ", v.ID, "2018-02-16", v.PublishDate)
				}
				expectedThumbnails := Thumbnails{
					Default:  ThumbnailDetail{URL: "https://i.ytimg.com/vi/UZxz9ot7y0Y/default.jpg", Width: 120, Height: 90},
					Medium:   ThumbnailDetail{URL: "https://i.ytimg.com/vi/UZxz9ot7y0Y/mqdefault.jpg", Width: 320, Height: 180},
					High:     ThumbnailDetail{URL: "https://i.ytimg.com/vi/UZxz9ot7y0Y/hqdefault.jpg", Width: 480, Height: 360},
					Standard: ThumbnailDetail{URL: "https://i.ytimg.com/vi/UZxz9ot7y0Y/sddefault.jpg", Width: 640, Height: 480},
					Maxres:   ThumbnailDetail{URL: "https://i.ytimg.com/vi/UZxz9ot7y0Y/maxresdefault.jpg", Width: 1280, Height: 720},
				}
				if !reflect.DeepEqual(v.Thumbnails, expectedThumbnails) {
					t.Errorf("id=%s: thumbnails expected %v, got %v ", v.ID, expectedThumbnails, v.Thumbnails)
				}
			}
		}
	}

}
func TestRelated(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock.NewMockClient(ctrl)

	apiKey := "foobar"
	req, errReq := http.NewRequest("GET", fmt.Sprintf(relatedSearchURL, "RiCql90xh7Q", 3, apiKey), nil)
	if errReq != nil {
		t.Fatal(errReq)
	}
	// request for related video id search
	c.EXPECT().Do(req).Return(returnFileAsResponse(relatedResultJSONPath))

	// request for relted video details
	req, errReq = http.NewRequest("GET", fmt.Sprintf(videoInfoURL, "lhu8HWc9TlA,mc7GUZinTD0,6qptaGpilE0", apiKey), nil)
	c.EXPECT().Do(req).Return(returnFileAsResponse(relatedDetailsJSONPath))

	DefaultVideoClient.apiKey = apiKey
	DefaultVideoClient.client = c

	// check related result
	if videos, err := DefaultVideoClient.Related("RiCql90xh7Q", 3); err != nil {
		t.Errorf("related failed, %s", err)
	} else {
		// check only number of obtained videos and test one video info details
		if len(videos) != 3 {
			t.Errorf("number of videos expected %d, got %d", 3, len(videos))
		}
		for _, v := range videos {
			// check video info details for id = lhu8HWc9TlA
			if v.ID == "lhu8HWc9TlA" {
				if v.Title != "Violet Evergarden - COMPLETE ALBUM - [ ヴァイオレット・エヴァーガーデン ] [BGM]" {
					t.Errorf("id=%s: title expected %s, got %s ", v.ID, "Violet Evergarden - COMPLETE ALBUM - [ ヴァイオレット・エヴァーガーデン ] [BGM]", v.Title)
				}
				if v.Duration != time.Hour*1+time.Minute*46+time.Second*52 {
					t.Errorf("id=%s: duration expected %s, got %s ", v.ID, time.Hour*1+time.Minute*46+time.Second*52, v.Duration)
				}
				if v.ViewCount != 9195 {
					t.Errorf("id=%s: view count expected %d, got %d ", v.ID, 9195, v.ViewCount)
				}
				if v.PublishDate != "2018-03-28" {
					t.Errorf("id=%s: publish date expected %s, got %s ", v.ID, "2018-03-28", v.PublishDate)
				}
				expectedThumbnails := Thumbnails{
					Default:  ThumbnailDetail{URL: "https://i.ytimg.com/vi/lhu8HWc9TlA/default.jpg", Width: 120, Height: 90},
					Medium:   ThumbnailDetail{URL: "https://i.ytimg.com/vi/lhu8HWc9TlA/mqdefault.jpg", Width: 320, Height: 180},
					High:     ThumbnailDetail{URL: "https://i.ytimg.com/vi/lhu8HWc9TlA/hqdefault.jpg", Width: 480, Height: 360},
					Standard: ThumbnailDetail{URL: "https://i.ytimg.com/vi/lhu8HWc9TlA/sddefault.jpg", Width: 640, Height: 480},
					Maxres:   ThumbnailDetail{},
				}
				if !reflect.DeepEqual(v.Thumbnails, expectedThumbnails) {
					t.Errorf("id=%s: thumbnails expected %v, got %v ", v.ID, expectedThumbnails, v.Thumbnails)
				}
			}
		}
	}

}

// returnFileAsResponse open file and return its reader object wrapped by http.Response
func returnFileAsResponse(filepath string) (*http.Response, error) {
	f, errOpen := os.Open(filepath)
	if errOpen != nil {
		return nil, errOpen
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       f,
	}, nil
}
