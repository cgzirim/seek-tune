package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"song-recognition/utils"

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
		os.Exit(1)
	}

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
		erase(SONGS_DIR)
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
		os.Exit(1)
	}
}
