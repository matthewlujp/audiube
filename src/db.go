// audios info and users info are stored in MongoDB.
// There are audios collection and users collection.
//
//
// A document in audios collection may contain,
// =============================================
//   - video id
//   - video info (youtube.Video)
//   - url to segment file of HLS
// =============================================
//
// A document in users collection may contain,
// =============================================
//   - uid
//   - email
//   - hashed password
// =============================================

package main

import (
	"errors"
	"fmt"
	"os"
	"sync"

	youtube "github.com/matthewlujp/audiube/src/youtube_data_v3"
	"gopkg.in/mgo.v2"
)

const (
	dbName              = "audiube"
	userCollectionName  = "users"
	audioCollectionName = "audios"
)

var (
	session              *mgo.Session
	onceDial             sync.Once
	closeLock            sync.Locker
	supportedCollections = map[string]struct{}{userCollectionName: struct{}{}, audioCollectionName: struct{}{}}
)

type videoDoc struct {
	VideoID        string
	VideoInfo      youtube.Video
	SegmentFileURL string
}

type userDoc struct {
	uid            string
	email          string
	hashedPassword string
}

// getSession returns a MongoDB session as a singleton.
// Handlers are supposed to obtain MongoDB session via this method, not directly.
func getSession() (*mgo.Session, error) {
	var errDial error
	// connect to db only once
	onceDial.Do(func() {
		// check MONGO_URI variable
		mongoURL := os.Getenv("MONGO_URI")
		if mongoURL == "" {
			mongoURL = "mongodb://localhost:27017" // default mongodb uri
		}
		// get session
		session, errDial = mgo.Dial(mongoURL)
		if errDial != nil {
			session = nil
		}
	})

	if errDial != nil {
		return nil, errors.New("dial to MongoDB has been failed")
	}
	if session == nil {
		return nil, errors.New("empty session, probably it has already been closed")
	}
	return session, nil
}

// getCollection returns either audios collection or users collection.
// If it cannot obtain a designated collection, since failure in session acquisition or unsupported collection, returns nil and error.
func getCollection(collectionName string) (*mgo.Collection, error) {
	s, errSession := getSession()
	if errSession != nil {
		return nil, errSession
	}

	// reject unsupported collection name
	if _, ok := supportedCollections[collectionName]; !ok {
		return nil, fmt.Errorf("collection %s not supported", collectionName)
	}

	return s.DB(dbName).C(collectionName), nil
}

// closeSession closes session if not nil
func closeSession() {
	closeLock.Lock() // lock to prevent being closed several times
	defer closeLock.Unlock()
	if session != nil {
		session.Close()
	}
	session = nil
}
