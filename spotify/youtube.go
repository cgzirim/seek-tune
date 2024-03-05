package spotify

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const developerKey = "AIzaSyC3nBFKqudeMItXnYKEeOUryLKhXnqBL7M"

// https://github.com/BharatKalluri/spotifydl/blob/v0.1.0/src/youtube.go
func VideoID(spTrack Track) (string, error) {
	service, err := youtube.NewService(context.TODO(), option.WithAPIKey(developerKey))
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
		return "", err
	}

	// Video category ID 10 is for music videos
	query := fmt.Sprintf("'%s' %s %s", spTrack.Title, spTrack.Artist, spTrack.Album) /* example: 'Lovesong' The Cure Disintegration */
	call := service.Search.List([]string{"id", "snippet"}).Q(query).VideoCategoryId("10").Type("video")

	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
		return "", err
	}
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			return item.Id.VideoId, nil
		}
	}
	// TODO: Handle when the query returns no songs (highly unlikely since the query is coming from spotify though)
	return "", nil
}
