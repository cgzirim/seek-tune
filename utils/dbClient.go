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

func (db *DbClient) InsertChunkData(chunkfgp int64, chunkData interface{}) error {
	chunksCollection := db.client.Database("song-recognition").Collection("chunks")

	filter := bson.M{"fingerprint": chunkfgp}

	var result bson.M
	err := chunksCollection.FindOne(context.Background(), filter).Decode(&result)
	if err == nil {
		// If the fingerprint already exists, append the chunkData to the existing list
		// fmt.Println("DUPLICATE FINGERPRINT: ", chunkfgp)
		update := bson.M{"$push": bson.M{"chunkData": chunkData}}
		_, err := chunksCollection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return fmt.Errorf("error updating chunk data: %v", err)
		}
		return nil
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// If the document doesn't exist, insert a new document
	_, err = chunksCollection.InsertOne(context.Background(), bson.M{"fingerprint": chunkfgp, "chunkData": []interface{}{chunkData}})
	if err != nil {
		return fmt.Errorf("error inserting chunk data: %v", err)
	}

	return nil
}

type chunkData struct {
	SongName     string `bson:"songName"`
	SongArtist   string `bson:"songArtist"`
	BitDepth     int    `bson:"bitDepth"`
	Channels     int    `bson:"channels"`
	SamplingRate int    `bson:"samplingRate"`
	TimeStamp    string `bson:"timeStamp"`
}

func (db *DbClient) GetChunkData(chunkfgp int64) ([]primitive.M, error) {
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

	var listOfChunkData []primitive.M
	for _, data := range result["chunkData"].(primitive.A) {
		listOfChunkData = append(listOfChunkData, data.(primitive.M))
	}

	return listOfChunkData, nil
}
