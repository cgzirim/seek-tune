package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected 'find', 'download', or 'serve' subcommands")
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
	default:
		fmt.Println("Expected 'find', 'download', or 'serve' subcommands")
		os.Exit(1)
	}
}
