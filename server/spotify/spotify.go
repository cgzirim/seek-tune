package spotify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"os"
	"song-recognition/utils"

	"github.com/tidwall/gjson"
)

/* for playlists and albums */
type ResourceEndpoint struct {
	Limit, Offset, TotalCount, Requests int64
}

type Track struct {
	Title, Artist, Album string
	Artists              []string
	Duration             int
}

const (
	tokenURL          = "https://accounts.spotify.com/api/token"
	cachedTokenPath   = "token.json"
)

type credentials struct {
	ClientID     string
	ClientSecret string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type cachedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func loadCredentials() (*credentials, error) {
	clientID := utils.GetEnv("SPOTIFY_CLIENT_ID", "")
	clientSecret := utils.GetEnv("SPOTIFY_CLIENT_SECRET", "")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID or SPOTIFY_CLIENT_SECRET environment variables not set")
	}

	return &credentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}


func saveToken(token string, expiresIn int) error {
	ct := cachedToken{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	data, err := json.MarshalIndent(ct, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachedTokenPath, data, 0644)
}

func loadCachedToken() (string, error) {
	data, err := os.ReadFile(cachedTokenPath)
	if err != nil {
		return "", err
	}
	var ct cachedToken
	if err := json.Unmarshal(data, &ct); err != nil {
		return "", err
	}
	if time.Now().After(ct.ExpiresAt) {
		return "", errors.New("token expired")
	}
	return ct.Token, nil
}

func accessToken() (string, error) {
	// Try using cached token
	token, err := loadCachedToken()
	if err == nil {
		return token, nil
	}

	// Fallback: request a new token
	creds, err := loadCredentials()
	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", creds.ClientID)
	data.Set("client_secret", creds.ClientSecret)

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", errors.New("token request failed (have a look at credentials.json): " + string(body))
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}

	if err := saveToken(tr.AccessToken, tr.ExpiresIn); err != nil {
		return "", err
	}

	return tr.AccessToken, nil
}

/* requests to playlist/track endpoints */
func request(endpoint string) (int, string, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, "", fmt.Errorf("error on making the request")
	}

	bearer, err := accessToken()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get access token: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+bearer)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("error on getting response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("error on reading response: %w", err)
	}

	return resp.StatusCode, string(body), nil
}

func getID(url string) string {
	parts := strings.Split(url, "/")
	id := strings.Split(parts[4], "?")[0]
	return id
}

func isValidPattern(url, pattern string) bool {
	match, _ := regexp.MatchString(pattern, url)
	return match
}

func TrackInfo(url string) (*Track, error) {
	re := regexp.MustCompile(`open\.spotify\.com\/(?:intl-.+\/)?track\/([a-zA-Z0-9]{22})(\?si=[a-zA-Z0-9]{16})?`)
	matches := re.FindStringSubmatch(url)
	if len(matches) <= 2 {
		return nil, errors.New("invalid track URL")
	}
	id := matches[1]

	endpoint := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", id)
	statusCode, jsonResponse, err := request(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting track info: %w", err)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("non-200 status code: %d", statusCode)
	}

	var result struct {
		Name     string `json:"name"`
		Duration int    `json:"duration_ms"`
		Album    struct {
			Name string `json:"name"`
		} `json:"album"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
	}
	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		return nil, err
	}

	var allArtists []string
	for _, a := range result.Artists {
		allArtists = append(allArtists, a.Name)
	}

	return (&Track{
		Title:    result.Name,
		Artist:   allArtists[0],
		Artists:  allArtists,
		Album:    result.Album.Name,
		Duration: result.Duration / 1000,
	}).buildTrack(), nil
}


func PlaylistInfo(url string) ([]Track, error) {
	re := regexp.MustCompile(`open\.spotify\.com\/playlist\/([a-zA-Z0-9]{22})`)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 2 {
		return nil, errors.New("invalid playlist URL")
	}
	id := matches[1]

	var allTracks []Track
	offset := 0
	limit := 100

	for {
		endpoint := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=%d&limit=%d", id, offset, limit)
		statusCode, jsonResponse, err := request(endpoint)
		if err != nil {
			return nil, fmt.Errorf("request error: %w", err)
		}
		if statusCode != 200 {
			return nil, fmt.Errorf("non-200 status: %d", statusCode)
		}

		var result struct {
			Items []struct {
				Track struct {
					Name     string `json:"name"`
					Duration int    `json:"duration_ms"`
					Album    struct {
						Name string `json:"name"`
					} `json:"album"`
					Artists []struct {
						Name string `json:"name"`
					} `json:"artists"`
				} `json:"track"`
			} `json:"items"`
			Total int `json:"total"`
		}
		if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			track := item.Track
			var artists []string
			for _, a := range track.Artists {
				artists = append(artists, a.Name)
			}
			allTracks = append(allTracks, *(&Track{
				Title:    track.Name,
				Artist:   artists[0],
				Artists:  artists,
				Duration: track.Duration / 1000,
				Album:    track.Album.Name,
			}).buildTrack())
		}

		offset += limit
		if offset >= result.Total {
			break
		}
	}

	return allTracks, nil
}

func AlbumInfo(url string) ([]Track, error) {
	re := regexp.MustCompile(`open\.spotify\.com\/album\/([a-zA-Z0-9]{22})`)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 2 {
		return nil, errors.New("invalid album URL")
	}
	id := matches[1]

	endpoint := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=50", id)
	statusCode, jsonResponse, err := request(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting album info: %w", err)
	}
	if statusCode != 200 {
		return nil, fmt.Errorf("non-200 status: %d", statusCode)
	}

	var result struct {
		Items []struct {
			Name     string `json:"name"`
			Duration int    `json:"duration_ms"`
			Artists  []struct {
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(jsonResponse), &result); err != nil {
		return nil, err
	}

	var tracks []Track
	for _, item := range result.Items {
		var artists []string
		for _, a := range item.Artists {
			artists = append(artists, a.Name)
		}
		tracks = append(tracks, *(&Track{
			Title:    item.Name,
			Artist:   artists[0],
			Artists:  artists,
			Duration: item.Duration / 1000,
			Album:    "", // You can fetch full album info if needed
		}).buildTrack())
	}

	return tracks, nil
}


/* returns playlist/album slice of tracks */
func resourceInfo(url, resourceType, totalCount, itemList string) ([]Track, error) {
	id := getID(url)
	eConf := ResourceEndpoint{Limit: 400, Offset: 0}
	jsonResponse, err := jsonList(resourceType, id, eConf.Offset, eConf.Limit)
	if err != nil {
		return nil, err
	}

	eConf.TotalCount = gjson.Get(jsonResponse, totalCount).Int()

	if eConf.TotalCount < 1 {
		return nil, errors.New("hum, there are no tracks")
	}

	name := map[bool]string{true: gjson.Get(jsonResponse, "data.playlistV2.name").String(), false: gjson.Get(jsonResponse, "data.albumUnion.name").String()}[resourceType == "playlist"]
	fmt.Printf("Collecting tracks from '%s'...\n", name)
	time.Sleep(1 * time.Second)

	eConf.Requests = int64(math.Ceil(float64(eConf.TotalCount) / float64(eConf.Limit))) /* total of requests */
	var tracks []Track
	tracks = append(tracks, proccessItems(jsonResponse, resourceType)...)

	for i := 1; i < int(eConf.Requests); i++ {
		eConf.pagination()

		jsonResponse, err := jsonList(resourceType, id, eConf.Offset, eConf.Limit)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, proccessItems(jsonResponse, resourceType)...)
	}

	fmt.Println("Tracks collected:", len(tracks))
	return tracks, nil
}

/* gets JSON respond from playlist/album endpoints */
func jsonList(resourceType, id string, offset, limit int64) (string, error) {
	var endpointQuery string
	var endpoint string
	if resourceType == "playlist" {
		endpointQuery = EncodeParam(fmt.Sprintf(`{"uri":"spotify:playlist:%s","offset":%d,"limit":%d}`, id, offset, limit))
		endpoint = endpointQuery
	} else {
		endpointQuery = EncodeParam(fmt.Sprintf(`{"uri":"spotify:album:%s","locale":"","offset":%d,"limit":%d}`, id, offset, limit))
		endpoint = endpointQuery
	}

	statusCode, jsonResponse, err := request(endpoint)
	if err != nil {
		return "", fmt.Errorf("error getting tracks: %w", err)
	}

	if statusCode != 200 {
		return "", fmt.Errorf("received non-200 status code: %d", statusCode)
	}

	return jsonResponse, nil
}

func (t *Track) buildTrack() *Track {
	track := &Track{
		Title:    t.Title,
		Artist:   t.Artist,
		Artists:  t.Artists,
		Duration: t.Duration,
		Album:    t.Album,
	}

	return track
}

func (eConf *ResourceEndpoint) pagination() {
	eConf.Offset = eConf.Offset + eConf.Limit
}

/* constructs each Spotify track from JSON body (album/playlist) and returns a slice of tracks */
func proccessItems(jsonResponse, resourceType string) []Track {
	itemList := map[bool]string{true: "data.playlistV2.content.items", false: "data.albumUnion.tracks.items"}[resourceType == "playlist"]
	songTitle := map[bool]string{true: "itemV2.data.name", false: "track.name"}[resourceType == "playlist"]
	artistName := map[bool]string{true: "itemV2.data.artists.items.0.profile.name", false: "track.artists.items.0.profile.name"}[resourceType == "playlist"]
	albumName := map[bool]string{true: "itemV2.data.albumOfTrack.name", false: "data.albumUnion.name"}[resourceType == "playlist"]
	duration := map[bool]string{true: "itemV2.data.trackDuration.totalMilliseconds", false: "track.duration.totalMilliseconds"}[resourceType == "playlist"]

	var tracks []Track
	items := gjson.Get(jsonResponse, itemList).Array()

	for _, item := range items {
		durationInSeconds := int(item.Get(duration).Int()) / 1000

		track := &Track{
			Title:    item.Get(songTitle).String(),
			Artist:   item.Get(artistName).String(),
			Duration: durationInSeconds,
			Album:    map[bool]string{true: item.Get(albumName).String(), false: gjson.Get(jsonResponse, albumName).String()}[resourceType == "playlist"],
		}
		tracks = append(tracks, *track.buildTrack())
	}

	return tracks
}
