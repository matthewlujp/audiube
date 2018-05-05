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
	"fmt"
	"net/http"
	"os"

	youtube "github.com/matthewlujp/audiube/src/youtube_data_v3"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	dbName              = "audiube"
	userCollectionName  = "users"
	audioCollectionName = "audios"
	dbSessionKey        = "db-session"
)

var (
	supportedCollections = map[string]struct{}{userCollectionName: struct{}{}, audioCollectionName: struct{}{}}
	mongoURL             string
)

func init() {
	// get MongoDB URI from an environment variable
	mongoURL = os.Getenv("MONGO_URI")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27017" // default mongodb uri
	}
}

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

// getCollection returns either audios collection or users collection.
// If it cannot obtain a designated collection, since failure in session acquisition or unsupported collection, returns nil and error.
func getCollection(sess *mgo.Session, collectionName string) (*mgo.Collection, error) {
	// reject unsupported collection name
	if _, ok := supportedCollections[collectionName]; !ok {
		return nil, fmt.Errorf("collection %s not supported", collectionName)
	}

	return sess.DB(dbName).C(collectionName), nil
}

// getSegmentListFileURLFromDB seeks a url for segment file in db.
// Error is returned if no entry is found.
func getSegmentListFileURLFromDB(sess *mgo.Session, videoID string) (string, error) {
	// search in the audio collection
	var result videoDoc
	if err := sess.DB(dbName).C(audioCollectionName).Find(
		bson.M{"videoid": videoID, "segmentfileurl": bson.M{"$exists": true}},
	).One(&result); err != nil {
		return "", fmt.Errorf("error occurred while searching segment list file url in db, %s", err) // error is returned if no document is found
	}
	return result.SegmentFileURL, nil
}

// setSegmentListFileURL inserts or overwrites segment list file url of videoID
func setSegmentListFileURL(sess *mgo.Session, videoID, fileURL string) error {
	// insert only if document for videoID does not exist, otherwise, update only the url
	searchQuery := bson.M{"videoid": videoID}
	c := sess.DB(dbName).C(audioCollectionName) // audio collection
	var res videoDoc
	if err := c.Find(searchQuery).One(&res); err == nil {
		// document for videoID exists -> update it
		return c.Update(searchQuery, bson.M{"$set": bson.M{"segmentfileurl": fileURL}})
	}

	// no document found for videoID -> insert
	return c.Insert(&struct {
		VideoID        string
		SegmentFileURL string
	}{VideoID: videoID, SegmentFileURL: fileURL})
}

// withDB create a db session and register to a variable manager
// supposed to be wraped by withVars beforehand
func withDB(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create db session
		sess, err := mgo.Dial(mongoURL)
		if err != nil {
			http.Error(w, "failed to connect to db, plz contact admin", http.StatusInternalServerError)
			return
		}
		defer sess.Close()

		// set session to vManager
		vManager.set(r, dbSessionKey, sess)
		f(w, r)
	}
}
