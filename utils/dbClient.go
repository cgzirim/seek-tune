package utils

import (
	"context"
	"fmt"

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
		return nil, err
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

func (db *DbClient) TotalSongs() (int, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("existing-songs")
	total, err := existingSongsCollection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

func (db *DbClient) SongExists(key string) (bool, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("existing-songs")
	filter := bson.M{"_id": key}

	var result bson.M
	if err := existingSongsCollection.FindOne(context.Background(), filter).Decode(&result); err == nil {
		return true, nil
	} else if err != mongo.ErrNoDocuments {
		return false, fmt.Errorf("error querying registered songs: %v", err)
	}

	return false, nil
}

func (db *DbClient) RegisterSong(key string) error {
	existingSongsCollection := db.client.Database("song-recognition").Collection("existing-songs")
	_, err := existingSongsCollection.InsertOne(context.Background(), bson.M{"_id": key})
	if err != nil {
		return fmt.Errorf("error registering song: %v", err)
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
			return fmt.Errorf("error updating chunk data: %v", err)
		}
		return nil
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// If the document doesn't exist, insert a new document
	_, err = chunksCollection.InsertOne(context.Background(), bson.M{"fingerprint": chunkfgp, "chunkTags": []interface{}{chunkTag}})
	if err != nil {
		return fmt.Errorf("error inserting chunk data: %v", err)
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
		return nil, fmt.Errorf("error retrieving chunk data: %w", err)
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
		return nil, fmt.Errorf("error finding chunk: %v", err)
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
