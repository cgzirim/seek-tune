module wasm-fingerprint

go 1.23.0

toolchain go1.24.3

require song-recognition v0.0.0-00010101000000-000000000000

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.17.6 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/mdobak/go-xerrors v0.3.1 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.14.0 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace song-recognition => ../server
