package youtube

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	searchURL        = "https://www.googleapis.com/youtube/v3/search?part=id&type=video&q=%s&maxResults=%d&key=%s"                // provide keywords(joined using %2C), max results, api key
	relatedSearchURL = "https://www.googleapis.com/youtube/v3/search?part=id&relatedToVideoId=%s&type=video&maxResults=%d&key=%s" // provide videoID,  max results, api key
	videoInfoURL     = "https://www.googleapis.com/youtube/v3/videos?part=id,snippet,contentDetails,statistics&id=%s&key=%s"      // provide videoID, api key
)

// VideoClient interface is responsible of retreiving video related data from YouTube Data v3
// It should take care of,
//  * searching based on keywords
//  * retreiving related videos
//  * get info of a video with a specific id
//
// Implementation of this interface requires a YouTube API key
type VideoClient interface {
	Search(keywords []string, maxResults int) ([]Video, error)
	Related(id string, maxResults int) ([]Video, error)
	Get(id string) (*Video, error)
}

type impleVideoClient struct {
	apiKey string
	client Client
}

// DefaultVideoClient is the only instance exported as a implementation of VideoClient interface
var DefaultVideoClient = &impleVideoClient{client: DefaultClient}

func init() {
	// YouTube Data v3 API key is searched from environmental variable.
	// Hence, it should be set before running this program.
	key := os.Getenv("API_KEY")
	if key == "" {
		log.Print("environment variable API_KEY is not set")
	}
	DefaultVideoClient.apiKey = key
}

// Search returns detailed info of videos hit for given keywords.
// It takes a []string for keywords and maximum number of videos for the search result.
// This method returns a []Video and error if any.
// Under the hood, the method issues GET request to YouTube Data v3 API twice,
//   1. to retrieve video ids related to given keywords and
//   2. to retrieve detailed video info for ids obtained via the previous step
func (c *impleVideoClient) Search(keywords []string, maxResults int) ([]Video, error) {
	// search with keyword and obtain video ids
	reqURL := fmt.Sprintf(searchURL, strings.Join(keywords, ","), maxResults, c.apiKey)
	req, errBuildReq := http.NewRequest("GET", reqURL, nil)
	if errBuildReq != nil {
		return nil, fmt.Errorf("failed in searching, %s", errBuildReq)
	}
	// GET request to YouTube Data v3 API for searching video ids
	res, errReq := c.client.Do(req)
	if errReq != nil {
		return nil, fmt.Errorf("failed in searching, %s", errReq)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response for %s got status %s", reqURL, res.Status)
	}

	// extract ids from json response
	ids, errIDs := parseSearchResult(res.Body)
	if errIDs != nil {
		return nil, fmt.Errorf("failed in searching, %s", errIDs)
	}

	// retrieve detailed video info for obtained video ids
	reqURL = fmt.Sprintf(videoInfoURL, strings.Join(ids, ","), c.apiKey)
	req, errBuildReq = http.NewRequest("GET", reqURL, nil)
	if errBuildReq != nil {
		return nil, fmt.Errorf("failed in detailed info for searched videos, %s", errBuildReq)
	}
	// GET request to YouTube Data v3 API for detailed info of each video id
	res, errReq = c.client.Do(req)
	if errReq != nil {
		return nil, fmt.Errorf("failed in detailed info for searched videos, %s", errReq)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response for %s got status %s", reqURL, res.Status)
	}

	// extract necessary data from returned json
	return parseVideosDetails(res.Body)
}

func (c *impleVideoClient) Related(id string, maxResults int) ([]Video, error) {
	return nil, errors.New("not implemented")
}

func (c *impleVideoClient) Get(id string) (*Video, error) {
	return nil, errors.New("not implemented")
}
