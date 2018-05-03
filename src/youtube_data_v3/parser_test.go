package youtube

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestParseSearchResult(t *testing.T) {
	f, errOpen := os.Open(searchResultJSONPath)
	if errOpen != nil {
		t.Fatal(errOpen)
	}
	defer f.Close()

	if ids, err := parseSearchResult(f); err != nil {
		t.Errorf("failed parsing search result, %s", err)
	} else {
		idSet := make(map[string]struct{})
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
		expectedSet := make(map[string]struct{})
		for _, id := range []string{"UZxz9ot7y0Y", "Ag4DR-L_TlM", "nzdDUg5R_IQ"} {
			expectedSet[id] = struct{}{}
		}
		if !reflect.DeepEqual(idSet, expectedSet) {
			t.Errorf("video ids expected %v, got %v", expectedSet, idSet)
		}
	}
}

func TestParseVideosDetails(t *testing.T) {
	f, errOpen := os.Open(searchDetailsJSONPath)
	if errOpen != nil {
		t.Fatal(errOpen)
	}
	defer f.Close()

	if videos, err := parseVideosDetails(f); err != nil {
		t.Errorf("failed parsing videos details, %s", err)
	} else {
		// check number of videos
		if len(videos) != 3 {
			t.Errorf("number of videos is wrong, expected %d, got %d", 3, len(videos))
		}

		// check one of the videos
		for _, v := range videos {
			if v.ID == "UZxz9ot7y0Y" {
				// only pickup title and default thumbnail url and check
				if v.Title != "Alisson Shore - Violet Ft. JMakata, Colt" {
					t.Errorf("video id %s title expected %s, got %s", "UZxz9ot7y0Y", "Alisson Shore - Violet Ft. JMakata, Colt", v.Title)
				}
				if v.Duration != time.Minute*4+time.Second*28 {
					t.Errorf("video id %s duration expected %s, got %s", "UZxz9ot7y0Y", time.Minute*4+time.Second*28, v.Duration)
				}
				if v.Thumbnails.Default.URL != "https://i.ytimg.com/vi/UZxz9ot7y0Y/default.jpg" {
					t.Errorf("video id %s thumbnail expected %s, got %s", "UZxz9ot7y0Y", "https://i.ytimg.com/vi/UZxz9ot7y0Y/default.jpg", v.Thumbnails.Default.URL)
				}
			}
		}
	}

}
