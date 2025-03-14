package db

import (
	"fmt"
	"song-recognition/models"
	"song-recognition/utils"
)

type DBClient interface {
	Close() error
	StoreFingerprints(fingerprints map[uint32]models.Couple) error
	GetCouples(addresses []uint32) (map[uint32][]models.Couple, error)
	TotalSongs() (int, error)
	RegisterSong(songTitle, songArtist, ytID string) (uint32, error)
	GetSong(filterKey string, value interface{}) (Song, bool, error)
	GetSongByID(songID uint32) (Song, bool, error)
	GetSongByYTID(ytID string) (Song, bool, error)
	GetSongByKey(key string) (Song, bool, error)
	DeleteSongByID(songID uint32) error
	DeleteCollection(collectionName string) error
}

type Song struct {
	Title     string
	Artist    string
	YouTubeID string
}

var DBtype = utils.GetEnv("DB_TYPE", "sqlite") // Can be "sqlite" or "mongo"

func NewDBClient() (DBClient, error) {
	switch DBtype {
	case "mongo":
		var (
			dbUsername = utils.GetEnv("DB_USER")
			dbPassword = utils.GetEnv("DB_PASS")
			dbName     = utils.GetEnv("DB_NAME")
			dbHost     = utils.GetEnv("DB_HOST")
			dbPort     = utils.GetEnv("DB_PORT")

			dbUri = "mongodb://" + dbUsername + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName
		)
		if dbUsername == "" || dbPassword == "" {
			dbUri = "mongodb://localhost:27017"
		}
		return NewMongoClient(dbUri)

	case "sqlite":
		return NewSQLiteClient("db/db.sqlite3")

	default:
		return nil, fmt.Errorf("unsupported database type: %s", DBtype)
	}
}
