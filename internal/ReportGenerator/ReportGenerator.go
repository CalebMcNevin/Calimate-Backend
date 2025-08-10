package ReportGenerator

import (
	// "bytes"
	"context"
	// "fmt"
	"io"
	"log"
	"os"
	// "text/template"

	"github.com/ollama/ollama/api"
)

func GenerateReport(transcriptPath string, outputPath string) {
	ctx := context.Background()

	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalf("failed to create Ollama client: %v", err)
	}

	// read transcription from `transcript.txt`
	transcript, err := os.ReadFile(transcriptPath)
	if err != nil {
		log.Fatalf("Failed to read transcript file: %v", err)
	}

	model := "lawnqc"
	prompt := string(transcript)

	// Open output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer outputFile.Close()

	// Stream to both stdout and file
	// writer := io.MultiWriter(os.Stdout, outputFile)
	writer := io.MultiWriter(outputFile)

	// Stream and print each chunk as it arrives
	err = client.Generate(ctx, &api.GenerateRequest{
		Model:     model,
		Prompt:    prompt,
		KeepAlive: &api.Duration{Duration: 0},
	}, func(resp api.GenerateResponse) error {
		_, err := writer.Write([]byte(resp.Response))
		return err
	})

	if err != nil {
		log.Fatalf("failed to stream response: %v", err)
	}
}
