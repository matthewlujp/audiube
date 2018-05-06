package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	youtube "github.com/matthewlujp/audiube/src/youtube_data_v3"
	mgo "gopkg.in/mgo.v2"
)

const (
	segmentListFilename = "audio.m3u8"
	segmentFilename     = "segment%04d.ts"
)

// handleWithLogging wraps a handler
func handleWithLogging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("Request URL:%s, Origin:%s, Content-Type:%s", r.RequestURI, r.RemoteAddr, r.Header.Get("Content-Type"))
		f(w, r)
	}
}

// allow request from other domain (front view program)
func handleWithCrossOriginRequest(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
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
// TODO: make a wrapper to alter segment list files
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	requestPath, errParse := parsePath(r.RequestURI)
	if errParse != nil {
		http.Error(w, errParse.Error(), http.StatusInternalServerError)
	}
	filePath := path.Join(*staticDirectory, requestPath.id)
	logger.Print("static file path ", filePath)

	f, errOpen := os.Open(filePath)
	if errOpen != nil {
		http.Error(w, errOpen.Error(), http.StatusInternalServerError)
		return
	}

	if filepath.Ext(filePath) == ".m3u8" {
		// if a request is for segment list file (.m3u8), insert "#EXT-X-START:0\n"
		w.Header().Set("Content-Type", "application/x-mpegurl")

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// insert "#EXT-X-START:0\n" after "#EXTM3U\n"
		originalList := buf.Bytes()
		startInstruction := []byte("#EXT-X-START:0\n")
		insertedList := make([]byte, 0, len(originalList)+len(startInstruction))
		insertedList = append(insertedList, originalList[:8]...) // len("#EXTM3U\n") = 8
		insertedList = append(insertedList, startInstruction...)
		insertedList = append(insertedList, originalList[8:]...)

		if _, err := f.Write(insertedList); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		w.Header().Set("Content-Type", name2ContentType(requestPath.id))
		if _, err := io.Copy(w, f); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
// URL is something like "static/streams/:videoID/audio.m3u8", since the program servers contents under static directory if requested.
// Embedding this url into video tag works.
// 3 cases to deal with,
//   * segment file url of a given video id exists in db -> respond with the url
//   * segment file url not in db but has already been created in the streams folder -> build url and respond with it, as well as registering it on db
//   * segment file has not been created -> download video, convert it using FFmpeg to HLS, respond with the url as soon as segment file (.m3u8) is created, and register the url on db
// This handler is supposed to be wraped by withVars and withDB.
func streamsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: path parsing should be done by a wrapper
	// TODO: graceful shutdown
	// TODO: stop and delete unsuccessful vedeo download
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

	// define response body json format
	// {
	// 	id: %s,
	// 	segment_list_file_url: %s
	// }
	resultFormat := ("{\"id\":\"%s\",\"segment_list_file_url\":\"%s\"}")

	// segment file exists and respond to the request with it
	if dbSess, ok := vManager.get(r, dbSessionKey).(*mgo.Session); ok {
		// got session
		if fileURL, err := getSegmentListFileURLFromDB(dbSess, pp.id); err == nil {
			// got url for segment list file
			if _, err := io.Copy(w, strings.NewReader(fmt.Sprintf(resultFormat, pp.id, fileURL))); err == nil {
				// writing done
				w.WriteHeader(http.StatusOK)
				return
			}
			// failed writing
			http.Error(
				w,
				fmt.Sprintf("error while writing url into response writer, %s", err),
				http.StatusInternalServerError,
			)
			return
		}
	}

	// search in static/streams directory
	// segment list file and video segments are stored in a directory whose name is a video id
	segmentListFilePath := path.Join(hlsSaveDirPath(pp.id), segmentListFilename)
	if _, err := os.Stat(segmentListFilePath); !os.IsNotExist(err) {
		// segment list file for videoID exists
		// respond with the url
		if _, err := io.Copy(w, strings.NewReader(fmt.Sprintf(resultFormat, pp.id, segmentListFilePath))); err != nil {
			http.Error(w, "error while writing url into response writer, "+err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)

		// register the url on db
		if sess, ok := vManager.get(r, dbSessionKey).(*mgo.Session); ok {
			if err := setSegmentListFileURL(sess, pp.id, segmentListFilePath); err != nil {
				logger.Printf("error while registering new segment list file url on db, %s", err)
			}
		}
		return
	}

	// download video, transcode with FFmpeg, and respond with a newly created segment list file url
	// download: use chunk fetch (goroutine)
	// FFmpeg: successively start transcoding from fetch data (goroutine)
	// respond: receive message from a goroutine where transcoding runs, respond to request and register segment list file url to db
	var errFetch error
	segmentListFilePath, errFetch = fetchVideAndBuildHLS(pp.id) // errorChannel notify result of transcode conducted in another goroutine
	if errFetch != nil {
		http.Error(w, errFetch.Error(), http.StatusInternalServerError)
	}

	// respond to a request after a while for segment list file creation
	time.Sleep(50 * time.Microsecond)
	if _, err := io.Copy(w, strings.NewReader(fmt.Sprintf(resultFormat, pp.id, segmentListFilePath))); err != nil {
		http.Error(w, "failed to write result into RewponseWriter "+err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK) // registration on DB will be done in next call for this vide id
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
