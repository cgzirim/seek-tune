package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"song-recognition/utils"

	"github.com/joho/godotenv"
	"github.com/mdobak/go-xerrors"
)

func main() {
	err := utils.CreateFolder("tmp")
	if err != nil {
		logger := utils.GetLogger()
		err := xerrors.New(err)
		ctx := context.Background()
		logger.ErrorContext(ctx, "Failed create tmp dir.", slog.Any("error", err))
	}

	err = utils.CreateFolder(SONGS_DIR)
	if err != nil {
		err := xerrors.New(err)
		logger := utils.GetLogger()
		ctx := context.Background()
		logMsg := fmt.Sprintf("failed to create directory %v", SONGS_DIR)
		logger.ErrorContext(ctx, logMsg, slog.Any("error", err))
	}

	if len(os.Args) < 2 {
		fmt.Println("Expected 'find', 'download', 'erase', 'save', or 'serve' subcommands")
		fmt.Println("\nUsage examples:")
		fmt.Println("  find <path_to_wav_file>")
		fmt.Println("  download <spotify_url>")
		fmt.Println("  erase [db | all]  (default: db)")
		fmt.Println("  save [-f|--force] <path_to_file_or_dir>")
		fmt.Println("  serve [-proto <http|https>] [-p <port>]")
		os.Exit(1)
	}
	_ = godotenv.Load()

	switch os.Args[1] {
	case "find":
		if len(os.Args) < 3 {
			fmt.Println("Usage: main.go find <path_to_wav_file>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		find(filePath)
	case "download":
		if len(os.Args) < 3 {
			fmt.Println("Usage: main.go download <spotify_url>")
			os.Exit(1)
		}
		url := os.Args[2]
		download(url)
	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		protocol := serveCmd.String("proto", "http", "Protocol to use (http or https)")
		port := serveCmd.String("p", "5000", "Port to use")
		serveCmd.Parse(os.Args[2:])
		serve(*protocol, *port)
	case "erase":
		// Default is to clear only database (db mode)
		dbOnly := true
		all := false

		if len(os.Args) > 2 {
			subCmd := os.Args[2]
			switch subCmd {
			case "db":
				dbOnly = true
				all = false
			case "all":
				dbOnly = false
				all = true
			default:
				fmt.Println("Usage: main.go erase [db | all]")
				fmt.Println("  db  : only clear the database (default)")
				fmt.Println("  all : clear database and songs folder")
				os.Exit(1)
			}
		}

		erase(SONGS_DIR, dbOnly, all)
	case "save":
		indexCmd := flag.NewFlagSet("save", flag.ExitOnError)
		force := indexCmd.Bool("force", false, "save song with or without YouTube ID")
		indexCmd.BoolVar(force, "f", false, "save song with or without YouTube ID (shorthand)")
		indexCmd.Parse(os.Args[2:])
		if indexCmd.NArg() < 1 {
			fmt.Println("Usage: main.go save [-f|--force] <path_to_wav_file_or_dir>")
			os.Exit(1)
		}
		filePath := indexCmd.Arg(0)
		save(filePath, *force)
	default:
		fmt.Println("Expected 'find', 'download', 'erase', 'save', or 'serve' subcommands")
		fmt.Println("\nUsage examples:")
		fmt.Println("  find <path_to_wav_file>")
		fmt.Println("  download <spotify_url>")
		fmt.Println("  erase [db | all]  (default: db)")
		fmt.Println("  save [-f|--force] <path_to_file_or_dir>")
		fmt.Println("  serve [-proto <http|https>] [-p <port>]")
		os.Exit(1)
	}
}
