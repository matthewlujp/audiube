package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	staticDirectory *string
	serverPort      *int
	logFilePath     *string
	logger          *log.Logger
)

func init() {
	staticDirectory = flag.String("static", "./static", "path to a directory where static files such as index.html are")
	serverPort = flag.Int("port", 5001, "port to listen")
	logFilePath = flag.String("log", "", "log output file path")
}

// YouTube audio player service.
// This service deals with,
//   * searching
//   * conversion from video to audio
//   * successive distribution
func main() {
	if *logFilePath == "" {
		logger = log.New(os.Stdout, "http: ", log.LstdFlags)
	} else {
		f, err := os.Create(*logFilePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		logger = log.New(f, "http: ", log.LstdFlags)
	}

	http.HandleFunc("/", handleWithLogging(indexHandler))
	http.HandleFunc("/static/", handleWithLogging(staticFileHandler))

	s := &http.Server{
		Addr:     fmt.Sprintf(":%d", *serverPort),
		Handler:  http.DefaultServeMux,
		ErrorLog: logger,
	}
	logger.Printf("audiube server start listening on port %d", *serverPort)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

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

// GET /videos?q=hoge
// {
// 	"hit_count": "30",
// 	"videos": [
// 		{
// 			"id": "a30jvlkjs03",
// 			"title": "hoge",
// 			"duration": "PT1H43M31S",
// 			"view_count": "136421",
// 			"publish_data": "2018-03-29",
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

// GET /videos?relatedID=a30jvlkjs03
// {
// 	"hit_count": "30",
// 	"videos": [
// 		{
// 			"id": "alskjeo93-s",
// 			"title": "hoge",
// 			"duration": "PT1H43M31S",
// 			"view_count": "136421",
// 			"publish_data": "2018-03-29",
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

// GET /videos/:id
// {
// 	"id": "a30jvlkjs03",
// 	"title": "hoge",
// 	"duration": "PT1H43M31S",
// 	"view_count": "136421",
// 	"publish_data": "2018-03-29",
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
// ====================================================================================================

// ====================================================================================================
// Resource: streams
// Desc: Returns a url of HLS segment list file

// GET /streams/:id
// {
// 	"id": "a30jvlkjs03",
// 	"segment_list_file_url": "/.../a30jvlkjs03/audio.m3u8",  <- relative path from index url
// }
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
// 				"publish_data": "2018-03-29",
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
