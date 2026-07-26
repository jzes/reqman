package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jzes/reqman/internal/adapter/requestfile"
	"github.com/jzes/reqman/internal/adapter/requesthttp"
	"github.com/jzes/reqman/internal/application"
	"github.com/jzes/reqman/internal/presentation"
)

func main() {
	requestsDir := "."
	if len(os.Args) > 1 {
		requestsDir = os.Args[1]
	}

	parsers := map[string]requestfile.Parser{
		requestfile.CurlFileExtension: requestfile.ParseCurlRequest,
	}
	fileSource := requestfile.NewFileSource(parsers)
	httpClient := requesthttp.NewClient()
	loader := application.NewRequestLoader(fileSource)
	writer := application.NewRequestWriter(fileSource, requestsDir)
	doer := application.NewRequestDoer(httpClient)

	requests, err := loader.LoadFromDirectory(requestsDir)
	if err != nil {
		log.Printf("Error loading requests from %q: %v", requestsDir, err)
		os.Exit(1)
	}

	program := tea.NewProgram(
		presentation.NewScreen(requests, writer, doer),
		tea.WithAltScreen(),
	)
	if _, err := program.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}

}
