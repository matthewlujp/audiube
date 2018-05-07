package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
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
	http.HandleFunc("/static/", handleWithLogging(allowCORS(staticFileHandler)))
	http.HandleFunc("/videos/", handleWithLogging(allowCORS(setContentTypeJSON(videosHandler))))
	http.HandleFunc("/streams/", handleWithLogging(allowCORS(setContentTypeJSON(withVars(withDB(streamsHandler))))))

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
