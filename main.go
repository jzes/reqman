package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jzes/reqman/internal/adapter/requestcurl"
	"github.com/jzes/reqman/internal/adapter/requestfile"
	"github.com/jzes/reqman/internal/adapter/requesthttp"
	"github.com/jzes/reqman/internal/presentation"
)

func main() {
	requestsDir := "."
	if len(os.Args) > 1 {
		requestsDir = os.Args[1]
	}

	formats := map[string]requestfile.Format{
		requestcurl.Extension: {
			Parse:  requestcurl.Parse,
			Format: requestcurl.Format,
		},
	}

	fileSource := requestfile.NewFileSource(formats)
	httpClient := requesthttp.NewClient()
	writer := requestfile.NewFileWriter(fileSource, requestsDir)

	requestPaths, err := fileSource.List(requestsDir)
	if err != nil {
		log.Printf("Error listing requests from %q: %v", requestsDir, err)
		os.Exit(1)
	}

	program := tea.NewProgram(
		presentation.NewScreen(requestPaths, writer, httpClient, fileSource),
		tea.WithAltScreen(),
	)
	if _, err := program.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

}
