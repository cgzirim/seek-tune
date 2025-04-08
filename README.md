<h1 align="center">SeekTune :musical_note:</h1>

<p align="center">
  <a href="https://drive.google.com/file/d/1I2esH2U4DtXHsNgYbUi4OL-ukV5i_1PI/view" target="_blank">
  <img src="https://github.com/user-attachments/assets/e4d01e9c-05cf-4f35-acbc-1e3cd79d1e00" 
       alt="screenshot" 
       width="500">
</a>
</p>

<p align="center"><a href="https://drive.google.com/file/d/1I2esH2U4DtXHsNgYbUi4OL-ukV5i_1PI/view" target="_blank">Demo in Video</a></p>

## Description 🎼
SeekTune is an implementation of Shazam's song recognition algorithm based on insights from these [resources](#resources--card_file_box). It integrates Spotify and YouTube APIs to find and download songs.

[//]: # (## Current Limitations
While the algorithm works excellently in matching a song with its exact file, it doesn't always find the right match from a recording. However, this project is still a work in progress. I'm hopeful about making it work, but I could definitely use some help :slightly_smiling_face:.   
Additionally, it currently only supports song files in WAV format.
)

## Installation :desktop_computer:
### Prerequisites
- Golang: [Install Golang](https://golang.org/dl/)
- FFmpeg: [Install FFmpeg](https://ffmpeg.org/download.html)
- NPM: To run the client (frontend).

### Steps
📦 Clone the repository:
```
git clone https://github.com/cgzirim/seek-tune.git
cd seek-tune
```
#### 🐳 Set Up with Docker
Prerequisites: [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
1. Build and run the application:
   ```Bash
   docker-compose up --build
   ```
   Visit the app at http://localhost:8080
2. To stop the application:
   ```Bash
   docker-compose down
   ```
#### 💻 Set Up Natively
Install dependencies for the backend
```
cd server
go get ./...
```
Install dependencies for the client
```
cd client
npm install
```

## Usage (Native Setup) :bicyclist:

#### ▸ Start the Client App 🏃‍♀️‍➡️ 
```
# Assuming you're in the client directory:

npm start
```
#### ▸ Start the Backend App 🏃‍♀️ 
In a separate terminal window:
```
cd server
go run *.go serve [-proto <http|https> (default: http)] [-port <port number> (default: 5000)]
```
#### ▸ Download a Song 📥 
Note: A link from Spotify's mobile app won't work. You can copy the link from either the desktop or web app.
```
go run *.go download <https://open.spotify.com/.../...>
```  
#### ▸ Save local songs to DB (supports all audio formats) 🗃️   
```
go run *.go save [-f|--force] <path_to_song_file_or_dir_of_songs>
```
The `-f` or `--force` flag allows saving the song even if a YouTube ID is not found. Note that the frontend will not display matches without a YouTube ID.  
  
#### ▸ Find matches for a song/recording 🔎
```
go run *.go find <path-to-wav-file>
```
#### ▸ Delete fingerprints and songs 🗑️ 
```
go run *.go erase
```

## Example :film_projector:  
Download a song 
```
$ go run *.go download https://open.spotify.com/track/4pqwGuGu34g8KtfN8LDGZm?si=b3180b3d61084018
Getting track info...
Now, downloading track...
Fingerprints saved in MongoDB successfully
'Voilà' by 'André Rieu' was downloaded
Total tracks downloaded: 1
```

Find matches of a song
```
$ go run *.go find songs/Voilà\ -\ André\ Rieu.wav
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

## Database Options 👯‍♀️ 
This application uses SQLite as the default database, but you can switch to MongoDB if preferred.   

#### Using MongoDB
1. [Install MongoDB](https://www.mongodb.com/docs/manual/installation/)
2. Configure MongoDB Connection:  
   To connect to your MongoDB instance, set the following environment variables:

   * `DB_TYPE`: Set this to "mongo" to indicate using MongoDB.
   * `DB_USER`: The username for your MongoDB database.
   * `DB_PASS`: The password for your MongoDB database.
   * `DB_NAME`: The name of the MongoDB database you want to use.
   * `DB_HOST`: The hostname or IP address of your MongoDB server.
   * `DB_PORT`: The port number on which your MongoDB server is listening.

   **Note:** The database connection URI is constructed using the environment variables.  
   If the `DB_USER` or `DB_PASS` environment variables are not set, it defaults to connecting to `mongodb://localhost:27017`.

## Resources  :card_file_box:
- [How does Shazam work - Coding Geek](https://drive.google.com/file/d/1ahyCTXBAZiuni6RTzHzLoOwwfTRFaU-C/view) (main resource)
- [Song recognition using audio fingerprinting](https://hajim.rochester.edu/ece/sites/zduan/teaching/ece472/projects/2019/AudioFingerprinting.pdf)
- [How does Shazam work - Toptal](https://www.toptal.com/algorithms/shazam-it-music-processing-fingerprinting-and-recognition)
- [Creating Shazam in Java](https://www.royvanrijn.com/blog/2010/06/creating-shazam-in-java/)


## Author :black_nib:
- Chigozirim Igweamaka
  - Connect with me on [LinkedIn](https://www.linkedin.com/in/ichigozirim/).
  - Check out my other [GitHub](https://github.com/cgzirim) projects.
  - Follow me on [Twitter](https://twitter.com/cgzirim).
 
## License :lock:
This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
