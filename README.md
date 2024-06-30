<h1 align="center">NotShazam :musical_note:</h1>

[NotShazam](https://notshazam.vercel.app/) is an implementation of Shazam's song recognition algorithm based on insights from these resources. It integrates Spotify and YouTube APIs to find and download songs from the internet.

## Current Limitations
While the algorithm works excellently in matching a song with its exact file, it performs poorly in identifying the right match from a recording. However, this project is still a work in progress. I'm hopeful about making it work, but I could definitely use some help.   
Additionally, it currently only supports song files in WAV format.

## Installation :desktop_computer:
### Prerequisites
- Golang: [Install Golang](https://golang.org/dl/)
- FFmpeg: [Install FFmpeg](https://ffmpeg.org/download.html)
- MongoDB: [Install MongoDB](https://www.mongodb.com/docs/manual/installation/)
- NPM: To run the client (frontend).

### Steps
Clone the repository:
```
git clone https://github.com/cgzirim/song-recognition.git
```
Install dependencies for the backend
```
cd song-recognition
go get ./...
```
Install dependencies for the client
```
cd song-recognition/client
npm install
```

## Usage :bicyclist:
Start the Client App
```
cd client
npm start
```
Serve the Backend App
```
go run main.go serve [-proto <http|https>] [-port <port number>]
```
Download a Song
```
go run main.go download <https://open.spotify.com/.../...>
```
Find matches for a song/recording
```
go run main.go find <path-to-wav-file>
```
Delete fingerprints and songs
```
go run main.go erase
```

### Example :film_projector:
Download a song 
```
$ go run main.go download https://open.spotify.com/track/4pqwGuGu34g8KtfN8LDGZm?si=b3180b3d61084018
Getting track info...
Now, downloading track...
Fingerprints saved in MongoDB successfully
'Voilà' by 'André Rieu' was downloaded
Total tracks downloaded: 1
```

Find matches of a song
```
$ go run main.go find songs/Voilà\ -\ André\ Rieu.wav
Top 20 matches:
        - Voilà by André Rieu, score: 5390686.00
        - I Am a Child of God by One Voice Children's Choir, score: 2539.00
        - I Have A Dream by ABBA, score: 2428.00
        - SOS by ABBA, score: 2327.00
        - Sweet Dreams (Are Made of This) - Remastered by Eurythmics, score: 2213.00
        - The Winner Takes It All by ABBA, score: 2094.00
        - Sleigh Ride by One Voice Children's Choir, score: 2091.00
        - Believe by Cher, score: 2089.00
        - Knowing Me, Knowing You by ABBA, score: 1958.00
        - Gimme! Gimme! Gimme! (A Man After Midnight) by ABBA, score: 1941.00
        - Take A Chance On Me by ABBA, score: 1932.00
        - Don't Stop Me Now - Remastered 2011 by Queen, score: 1892.00
        - I Do, I Do, I Do, I Do, I Do by ABBA, score: 1853.00
        - Everywhere - 2017 Remaster by Fleetwood Mac, score: 1779.00
        - You Will Be Found by One Voice Children's Choir, score: 1664.00
        - J'Imagine by One Voice Children's Choir, score: 1658.00
        - When You Believe by One Voice Children's Choir, score: 1629.00
        - When Love Was Born by One Voice Children's Choir, score: 1484.00
        - Don't Stop Believin' (2022 Remaster) by Journey, score: 1465.00
        - Lay All Your Love On Me by ABBA, score: 1436.00

Search took: 856.386557ms

Final prediction: Voilà by André Rieu , score: 5390686.00
```
## Resources
- [How does Shazam work - Coding Geek](https://drive.google.com/file/d/1ahyCTXBAZiuni6RTzHzLoOwwfTRFaU-C/view) (main resource)
- [Song recognition using audio fingerprinting](https://hajim.rochester.edu/ece/sites/zduan/teaching/ece472/projects/2019/AudioFingerprinting.pdf)
- [How does Shazam work - Toptal](https://www.toptal.com/algorithms/shazam-it-music-processing-fingerprinting-and-recognition)
- [Creating Shazam in Java](https://www.royvanrijn.com/blog/2010/06/creating-shazam-in-java/)

## Author :black_nib:
- Chigozirim Igweamaka
  - Check out my other [GitHub](https://github.com/cgzirim) projects.
  - Connect with me on [LinkedIn](https://www.linkedin.com/in/chigozirim-igweamaka/).
  - Follow me on [Twitter](https://twitter.com/cgzirim).
 
