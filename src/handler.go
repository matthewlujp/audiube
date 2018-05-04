package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	youtube "github.com/matthewlujp/audiube/src/youtube_data_v3"
	"gopkg.in/mgo.v2/bson"
)

// handleWithLogging wraps a handler
func handleWithLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Request URL:%s, Origin:%s, Content-Type:%s", r.RequestURI, r.RemoteAddr, r.Header.Get("Content-Type"))
		f(w, r)
	}
}

// GET /
// index.html <- single page driven by React
// indexHandler provides index page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	indexPath := path.Join(*staticDirectory, "index.html")
	f, errOpen := os.Open(indexPath)
	if errOpen != nil {
		http.Error(w, errOpen.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handleWithContentTypeJSON sets content-type field to application/json
func handleWithContentTypeJSON(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		f(w, r)
	}
}

// GET /static/hoge
// Desc: static files such as css and js are placed under static directory and obtained through this handler
// staticFileHandler provides filename under static directory for /filename request
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	requestPath, errParse := parsePath(r.RequestURI)
	if errParse != nil {
		http.Error(w, errParse.Error(), http.StatusInternalServerError)
	}
	filePath := path.Join(*staticDirectory, requestPath.id)
	logger.Print("static file path", filePath)
	f, errOpen := os.Open(filePath)
	if errOpen != nil {
		http.Error(w, errOpen.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", name2ContentType(requestPath.id))
	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ====================================================================================================
// Resource: videos
// Desc: video information such as title, thumbnail, etc
// videoHandler switch handler according to following path or parameters
func videosHandler(w http.ResponseWriter, r *http.Request) {
	// parse request path
	pp, errParse := parsePath(r.URL.String())
	if errParse != nil {
		http.Error(w, fmt.Sprintf("parsing request %s failed, %s", r.URL.RawPath, errParse), http.StatusBadRequest)
		return
	}

	// switch according to parsed path
	switch {
	case pp.id != "": // /videos/id -> get info of video with id
		videoGetHandler(w, r)
	case pp.params.Get("q") != "": // /videos?q=hoge -> key words search
		searchHandler(w, r)
	case pp.params.Get("relatedToVideoId") != "": // /videos?relatedId=foo ->  search for related videos to id
		relatedSearchHandler(w, r)
	default:
		http.Error(w, fmt.Sprintf("unsupported request %s", r.URL.RawPath), http.StatusBadRequest)
	}
}

// GET /videos?q=hoge
// {
// 	"hit_count": "30",
// 	"videos": [
// 		{
// 			"id": "a30jvlkjs03",
// 			"title": "hoge",
// 			"duration": "PT1H43M31S",
// 			"view_count": "136421",
// 			"publish_date": "2018-03-29",
// 			"thumbnail": {
// 				"default": {
// 					"url": "https//..." ,
// 					"width": "320",
// 					"height": "180",
// 				}
// 				"medium": {
// 					"url": "https//..." ,
// 					"width": "480",
// 					"height": "360",
// 				}
// 				"standard": {
// 					"url": "https//..." ,
// 					"width": "640",
// 					"height": "480",
// 				}
// 			}
// 		},...
// 	]
// }
// If no query param q is found or empty, returns an empty json.
// Query param is separated by ",".
func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	qParams, ok := q["q"]
	// define result structure
	var resp struct {
		HitCount int             `json:"hit_count"`
		Videos   []youtube.Video `json:"videos"`
	}

	if !ok {
		// return json with empty videos
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// split qParams and search with the given keywords
	keywords := strings.Split(qParams[0], ",")
	videos, errSearch := youtube.DefaultVideoClient.Search(keywords, 50) // result videos are 50 at most
	if errSearch != nil {
		http.Error(w, errSearch.Error(), http.StatusInternalServerError)
		return
	}

	// write out result
	resp.HitCount = len(videos)
	resp.Videos = videos
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /videos?relatedToVideoId=a30jvlkjs03
// {
// 	"hit_count": "30",
// 	"videos": [
// 		{
// 			"id": "alskjeo93-s",
// 			"title": "hoge",
// 			"duration": "PT1H43M31S",
// 			"view_count": "136421",
// 			"publish_date": "2018-03-29",
// 			"thumbnail": {
// 				"default": {
// 					"url": "https//..." ,
// 					"width": "320",
// 					"height": "180",
// 				}
// 				"medium": {
// 					"url": "https//..." ,
// 					"width": "480",
// 					"height": "360",
// 				}
// 				"high": {
// 					"url": "https//..." ,
// 					"width": "640",
// 					"height": "480",
// 				}
// 			}
// 		},...
// 	]
// }
func relatedSearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	qParams, ok := q["relatedToVideoId"]
	// define result structure
	var resp struct {
		HitCount int             `json:"hit_count"`
		Videos   []youtube.Video `json:"videos"`
	}

	if !ok {
		// return json with empty videos
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// search related video with a given id
	videos, errSearch := youtube.DefaultVideoClient.Related(qParams[0], 50) // result videos are 50 at most
	if errSearch != nil {
		http.Error(w, errSearch.Error(), http.StatusInternalServerError)
		return
	}

	// write out result
	resp.HitCount = len(videos)
	resp.Videos = videos
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /videos/:id
// {
// 	"id": "a30jvlkjs03",
// 	"title": "hoge",
// 	"duration": "PT1H43M31S",
// 	"view_count": "136421",
// 	"publish_date": "2018-03-29",
// 	"thumbnail": {
// 		"default": {
// 			"url": "https//..." ,
// 			"width": "320",
// 			"height": "180",
// 		}
// 		"medium": {
// 			"url": "https//..." ,
// 			"width": "480",
// 			"height": "360",
// 		}
// 		"standard": {
// 			"url": "https//..." ,
// 			"width": "640",
// 			"height": "480",
// 		}
// 	}
// }
func videoGetHandler(w http.ResponseWriter, r *http.Request) {
	pp, _ := parsePath(r.URL.String()) // error has already been checked

	video, err := youtube.DefaultVideoClient.Get(pp.id) // result videos are 50 at most
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write out result
	if err := json.NewEncoder(w).Encode(video); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ====================================================================================================

// ====================================================================================================
// Resource: streams
// Desc: Returns a url of HLS segment list file

// GET /streams/:id
// {
// 	"id": "a30jvlkjs03",
// 	"segment_list_file_url": "/.../a30jvlkjs03/audio.m3u8",  <- relative path from index url
// }
//
// streamHandler respond to a stream request, i.e. request for HLS segment file
// 4 cases
//   * segment file url of a given video id exists in db -> respond with the url
//   * segment file url not in db but has already been created in the streams folder -> build url and respond with it, as well as registering it on db
//   * segment file has not been created -> download video, convert it using FFmpeg to HLS, respond with the url as soon as segment file (.m3u8) is created, and register the url on db
func streamsHandler(w http.ResponseWriter, r *http.Request) {
	// get video id from request path
	pp, errParse := parsePath(r.URL.String())
	if errParse != nil {
		http.Error(w, fmt.Sprintf("failed to parse request path %s, %s", r.URL, errParse), http.StatusBadRequest)
		return
	}
	if pp.id == "" {
		http.Error(w, "id empty", http.StatusBadRequest)
		return
	}

	if fileURL := getSegmentListFileURLFromDB(pp.id); fileURL != "" {
		// segment file exists and respond to the request with it
		result := struct {
			ID      string `json:"id"`
			FileURL string `json:"segment_list_file_url"`
		}{ID: pp.id, FileURL: fileURL}
		json.NewEncoder(w).Encode(result)
		w.WriteHeader(http.StatusOK)
		return
	}

	io.Copy(w, strings.NewReader(fmt.Sprintf("{\"id\":\"%s\",\"message\":\"video id not found...\"", pp.id)))
	w.WriteHeader(http.StatusBadRequest)
}

// getSegmentListFileURLFromDB looks for segment file url in db.
// In either case where no entry is found or an error occurs, empty string is returned
func getSegmentListFileURLFromDB(videoID string) string {
	c, errCollection := getCollection(audioCollectionName)
	if errCollection != nil {
		logger.Printf("error occurred while connecting to db, %s", errCollection)
		return ""
	}

	// search in a collection of db
	var result videoDoc
	if err := c.Find(bson.M{"videoid": videoID, "segmentfileurl": bson.M{"$exists": true}}).One(&result); err != nil {
		logger.Printf("error occurred while searching segment list file url in db, %s", err)
		return ""
	}
	if result.SegmentFileURL == "" { // url doesn't exit in db
		return ""
	}
	return result.SegmentFileURL
}

// ====================================================================================================

// ====================================================================================================
// Resource: users
// Desc: Manage user information (session management)

// GET /users/:id <- authentication is required
// {
//	"user": {
//		"name": "Matthew Ishige",
//		"email": "matthewlujp@gmail.com",
//		"uid": "laksa09lksj",
//	},
// 	"recently_played": {
// 		"count": "30",
// 		"videos": [
// 			{
// 				"id": "a30jvlkjs03",
// 				"title": "hoge",
// 				"duration": "PT1H43M31S",
// 				"view_count": "136421",
// 				"publish_date": "2018-03-29",
// 				"thumbnail": {
// 					"default": {
// 						"url": "https//..." ,
// 						"width": "320",
// 						"height": "180",
// 					}
// 					"medium": {
// 						"url": "https//..." ,
// 						"width": "480",
// 						"height": "360",
// 					}
// 					"standard": {
// 						"url": "https//..." ,
// 						"width": "640",
// 						"height": "480",
// 					}
// 				}
// 			},...
// 		]
// 	}
// }

// POST /users
// DATA
// {
// 	"name": "Matthew Ishige",
// 	"email": "matthewlujp@gmail.com",
// 	"passwd": "abcdefg",
// }
// RESPONSE
// {
// 	"message": "Welcome!",
//	"user": {
//		"name": "Matthew Ishige",
//		"email": "matthewlujp@gmail.com",
//		"uid": "laksa09lksj",
//	}
// }

// UPDATE /users
// DATA
// {
// 	"name": "Matthew Lu",
// }
// RESPONSE
// {
// 	"message": "Update succeeded."
//	"user": {
//		"name": "Matthew Ishige",
//		"email": "matthewlujp@gmail.com",
//		"uid": "laksa09lksj",
//	}
// }
// ====================================================================================================

// ====================================================================================================
// Resource: login
// Desc: Manage loging process. Session expire mechanism is necessary.

// POST /login
// DATA
// {
// 	"email": "matthewlujp@gmail.com",
// 	"passwd": "foobar",
// }
// RESPONSE
// {
// 	"message": "login succeeded",
// 	"user": {
// 		"name": "Matthew Ishige",
// 		"email": "matthewlujp@gmail.com",
// 		"uid": "laksa09lksj",
// 	}
// }
