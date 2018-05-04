package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/matthewlujp/gotube"
)

// fetchVideoAndBuildHLS is responsible for three tasks
//   1.download: use chunk fetch (goroutine)
//   2.FFmpeg: successively start transcoding from fetch data (goroutine)
//   3.respond: receive message from a goroutine where transcoding runs, respond to request and register segment list file url to db
// Returns,
//   * segment list file path
//   * error
func fetchVideAndBuildHLS(videoID string) (string, error) {
	// TODO: remove failed HLS file
	// download video
	stream, errSelect := selectStream(videoID)
	if errSelect != nil {
		return "", errSelect
	}
	dataChan, errDL := stream.SequentialChunkDownload(10 * time.Second) // conducted in a goroutine
	if errDL != nil {
		return "", errDL
	}

	// create diretory to save transcoded audio files
	if _, err := os.Stat(hlsSaveDirPath(videoID)); os.IsNotExist(err) {
		if errCreate := os.MkdirAll(hlsSaveDirPath(videoID), 0777); errCreate != nil {
			return "", fmt.Errorf("failed to create directory to save transcoded audio file, %s", errCreate)
		}
	}

	// prepare for FFmpeg transcode
	segmentListFilePath := path.Join(hlsSaveDirPath(videoID), segmentListFilename)
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", "pipe:0",
		"-codec:", "copy",
		"-vn",
		"-ss", "0",
		"-t", format(stream.Duration),
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		segmentListFilePath,
	)
	// get input pipeline for FFmpeg
	w, errStdin := cmd.StdinPipe()
	if errStdin != nil {
		return "", fmt.Errorf("failed to get stdin for FFmpeg command execution, %s", errStdin)
	}
	// send received data to input pipe of FFmpeg
	go func() {
		defer w.Close()
		for data := range dataChan {
			w.Write(data)
		}
	}()
	cmd.Start() // start transcoding by FFmpeg in background
	return segmentListFilePath, nil
}

// selectStream selects propper stream for videoID.
func selectStream(videoID string) (*gotube.Stream, error) {
	// prepare downloader and collect necessary info
	downloader, errBuild := gotube.NewDownloader(
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
	)
	if errBuild != nil {
		return nil, errBuild
	}
	if err := downloader.FetchStreams(); err != nil {
		return nil, err
	}

	// seek appropriate streams
	// policy:
	//   1. seek mp4 && audio
	//   2. seek mp4 && smallest resolution
	candidateStream := downloader.Streams[0]
	for _, stream := range downloader.Streams {
		// if current candidate is not mp4, force to exchange
		if candidateStream.Format != "mp4" {
			candidateStream = stream
		}

		if stream.Format != "mp4" {
			continue
		}

		if stream.MediaType == "audio" {
			// break if stream is mp4 and audio
			candidateStream = stream
			break
		}

		if r, err := resCmp(stream, candidateStream); err == nil && r < 0 {
			// change to smaller stream only if compare succeeds
			candidateStream = stream
		}
	}

	return candidateStream, nil
}

// "480p" -> 480
func resolution2Numeric(res string) (int, error) {
	return strconv.Atoi(strings.TrimSuffix(res, "p"))
}

// resCmp returns one of  -1, 0, or 1
//   -1 if stream1.Resolution < stream2.Resolution
//   0 if stream1.Resolution == stream2.Resolution
//   1 if stream1.Resolution > stream2.Resolution
func resCmp(stream1, stream2 *gotube.Stream) (int, error) {
	r1, err1 := resolution2Numeric(stream1.Resolution)
	if err1 != nil {
		return 0, err1
	}
	r2, err2 := resolution2Numeric(stream2.Resolution)
	if err2 != nil {
		return 0, err2
	}

	var cmpRes int
	switch {
	case r1 < r2:
		cmpRes = -1
	case r1 == r2:
		cmpRes = 0
	case r1 > r2:
		cmpRes = 1
	}
	return cmpRes, nil
}

// hlsSaveDirPath defines where to save HLS file
func hlsSaveDirPath(videoID string) string {
	return path.Join(*staticDirectory, "streams", videoID) // static/streams/videoID
}

// format build string form duration for ffmpeg option -t.
// time.Duration to "01:23:11"
func format(d time.Duration) string {
	h := int(math.Floor(d.Hours()))
	m := int(math.Floor(d.Minutes())) % 60
	s := int(math.Floor(d.Seconds())) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
