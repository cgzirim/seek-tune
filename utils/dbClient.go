package utils

import (
	"context"
	"fmt"
	"song-recognition/models"

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
				SongID:       itemMap["songID"].(string),
			}
			docTables = append(docTables, table)
		}
		tables[address] = docTables
	}

	return tables, nil
}

func (db *DbClient) TotalSongs() (int, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("existing-songs")
	total, err := existingSongsCollection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

func (db *DbClient) SongExists(songTitle, songArtist, ytID string) (bool, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("existing-songs")

	key := fmt.Sprintf("%s - %s", songTitle, songArtist)
	var filter bson.M

	if len(ytID) == 0 {
		filter = bson.M{"_id": key}
	} else {
		filter = bson.M{"ytID": ytID}
	}

	var result bson.M
	if err := existingSongsCollection.FindOne(context.Background(), filter).Decode(&result); err == nil {
		return true, nil
	} else if err != mongo.ErrNoDocuments {
		return false, fmt.Errorf("failed to retrieve registered songs: %v", err)
	}

	return false, nil
}

func (db *DbClient) RegisterSong(songTitle, songArtist, ytID string) error {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")

	// Create a compound unique index on ytID and key, if it doesn't already exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"ytID", 1}, {"key", 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := existingSongsCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return fmt.Errorf("failed to create unique index: %v", err)
	}

	// Attempt to insert the song with ytID and key
	key := fmt.Sprintf("%s - %s", songTitle, songArtist)
	_, err = existingSongsCollection.InsertOne(context.Background(), bson.M{"_id": key, "ytID": ytID})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("song with ytID or key already exists: %v", err)
		} else {
			return fmt.Errorf("failed to register song: %v", err)
		}
	}

	return nil
}

func (db *DbClient) InsertChunkTag(chunkfgp int64, chunkTag interface{}) error {
	chunksCollection := db.client.Database("song-recognition").Collection("chunks")

	filter := bson.M{"fingerprint": chunkfgp}

	var result bson.M
	err := chunksCollection.FindOne(context.Background(), filter).Decode(&result)
	if err == nil {
		// If the fingerprint already exists, append the chunkTag to the existing list
		// fmt.Println("DUPLICATE FINGERPRINT: ", chunkfgp)
		update := bson.M{"$push": bson.M{"chunkTags": chunkTag}}
		_, err := chunksCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("failed to update chunkTags: %v", err)
		}
		return nil
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// If the document doesn't exist, insert a new document
	_, err = chunksCollection.InsertOne(context.Background(), bson.M{"fingerprint": chunkfgp, "chunkTags": []interface{}{chunkTag}})
	if err != nil {
		return fmt.Errorf("failed to insert chunk tag: %v", err)
	}

	return nil
}

func (db *DbClient) GetChunkTags(chunkfgp int64) ([]primitive.M, error) {
	chunksCollection := db.client.Database("song-recognition").Collection("chunks")

	filter := bson.M{"fingerprint": chunkfgp}
	result := bson.M{}
	err := chunksCollection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve chunk tag: %w", err)
	}

	var listOfChunkTags []primitive.M
	for _, data := range result["chunkTags"].(primitive.A) {
		listOfChunkTags = append(listOfChunkTags, data.(primitive.M))
	}

	return listOfChunkTags, nil
}

func (db *DbClient) GetChunkTagForSong(songTitle, songArtist string) (bson.M, error) {
	chunksCollection := db.client.Database("song-recognition").Collection("chunks")

	filter := bson.M{
		"chunkTags": bson.M{
			"$elemMatch": bson.M{
				"songtitle":  songTitle,
				"songartist": songArtist,
			},
		},
	}

	var result bson.M
	if err := chunksCollection.FindOne(context.Background(), filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find chunk: %v", err)
	}

	var chunkTag map[string]interface{}
	for _, chunk := range result["chunkTags"].(primitive.A) {
		chunkMap, ok := chunk.(primitive.M)
		if !ok {
			continue
		}
		if chunkMap["songtitle"] == songTitle && chunkMap["songartist"] == songArtist {
			chunkTag = chunkMap
			break
		}
	}

	return chunkTag, nil
}
