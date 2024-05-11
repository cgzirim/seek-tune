package utils

import (
	"context"
	"errors"
	"fmt"
	"song-recognition/models"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbUri string = "mongodb://localhost:27017"

// DbClient represents a MongoDB client
type DbClient struct {
	client *mongo.Client
}

// NewDbClient creates a new instance of DbClient
func NewDbClient() (*DbClient, error) {
	clientOptions := options.Client().ApplyURI(dbUri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %d", err)
	}
	return &DbClient{client: client}, nil
}

// Close closes the underlying MongoDB client
func (db *DbClient) Close() error {
	if db.client != nil {
		return db.client.Disconnect(context.Background())
	}
	return nil
}

func (db *DbClient) StoreFingerprints(fingerprints map[uint32]models.Table) error {
	collection := db.client.Database("song-recognition").Collection("fingerprints")

	for address, table := range fingerprints {
		// Check if the address already exists in the database
		var existingDoc bson.M
		err := collection.FindOne(context.Background(), bson.M{"_id": address}).Decode(&existingDoc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				// If address doesn't exist, insert a new document
				doc := bson.M{
					"_id": address,
					"tables": []interface{}{
						bson.M{
							"anchorTimeMs": table.AnchorTimeMs,
							"songID":       table.SongID,
						},
					},
				}

				_, err := collection.InsertOne(context.Background(), doc)
				if err != nil {
					return fmt.Errorf("error inserting document: %s", err)
				}
			} else {
				return fmt.Errorf("error checking if document exists: %s", err)
			}
		} else {
			// If address exists, append the new table to the existing tables list

			_, err := collection.UpdateOne(
				context.Background(),
				bson.M{"_id": address},
				bson.M{"$push": bson.M{"tables": bson.M{"anchorTimeMs": table.AnchorTimeMs, "songID": table.SongID}}},
			)
			if err != nil {
				return fmt.Errorf("error updating document: %s", err)
			}
		}
	}

	return nil
}

func (db *DbClient) GetTables(addresses []uint32) (map[uint32][]models.Table, error) {
	collection := db.client.Database("song-recognition").Collection("fingerprints")

	tables := make(map[uint32][]models.Table)

	for _, address := range addresses {
		// Find the document corresponding to the address
		var result bson.M
		err := collection.FindOne(context.Background(), bson.M{"_id": address}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			}
			return nil, fmt.Errorf("error retrieving document for address %d: %s", address, err)
		}

		// Extract tables from the document and append them to the tables map
		var docTables []models.Table
		tableArray, ok := result["tables"].(primitive.A)
		if !ok {
			return nil, fmt.Errorf("tables field in document for address %d is not valid", address)
		}

		for _, item := range tableArray {
			itemMap, ok := item.(primitive.M)
			if !ok {
				return nil, fmt.Errorf("invalid table format in document for address %d", address)
			}

			table := models.Table{
				AnchorTimeMs: uint32(itemMap["anchorTimeMs"].(int64)),
				SongID:       uint32(itemMap["songID"].(int64)),
			}
			docTables = append(docTables, table)
		}
		tables[address] = docTables
	}

	return tables, nil
}

func (db *DbClient) TotalSongs() (int, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")
	total, err := existingSongsCollection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

func (db *DbClient) RegisterSong(songTitle, songArtist, ytID string) (uint32, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")

	// Create a compound unique index on ytID and key, if it doesn't already exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"ytID", 1}, {"key", 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := existingSongsCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return 0, fmt.Errorf("failed to create unique index: %v", err)
	}

	// Attempt to insert the song with ytID and key
	songID := GenerateUniqueID()
	key := GenerateSongKey(songTitle, songArtist)
	_, err = existingSongsCollection.InsertOne(context.Background(), bson.M{"_id": songID, "key": key, "ytID": ytID})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, fmt.Errorf("song with ytID or key already exists: %v", err)
		} else {
			return 0, fmt.Errorf("failed to register song: %v", err)
		}
	}

	return songID, nil
}

type Song struct {
	Title     string
	Artist    string
	YouTubeID string
}

const FILTER_KEYS = "_id | ytID | key"

func (db *DbClient) GetSong(filterKey string, value interface{}) (s Song, songExists bool, e error) {
	if !strings.Contains(FILTER_KEYS, filterKey) {
		return Song{}, false, errors.New("invalid filter key")
	}

	songsCollection := db.client.Database("song-recognition").Collection("songs")
	var song bson.M

	filter := bson.M{filterKey: value}

	err := songsCollection.FindOne(context.Background(), filter).Decode(&song)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Song{}, false, nil
		}
		return Song{}, false, fmt.Errorf("failed to retrieve song: %v", err)
	}

	ytID := song["ytID"].(string)
	title := strings.Split(song["key"].(string), "---")[0]
	artist := strings.Split(song["key"].(string), "---")[1]

	songInstance := Song{title, artist, ytID}

	return songInstance, true, nil
}

func (db *DbClient) GetSongByID(songID uint32) (Song, bool, error) {
	return db.GetSong("_id", songID)
}

func (db *DbClient) GetSongByYTID(ytID string) (Song, bool, error) {
	return db.GetSong("ytID", ytID)
}

func (db *DbClient) GetSongByKey(key string) (Song, bool, error) {
	return db.GetSong("key", key)
}

func (db *DbClient) DeleteSongByID(songID uint32) error {
	songsCollection := db.client.Database("song-recognition").Collection("songs")

	filter := bson.M{"_id": songID}

	_, err := songsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete song: %v", err)
	}

	return nil
}
