package uploads

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"qc_api/internal/ReportGenerator"
	"qc_api/internal/jobqueue"
)

type UploadResponse struct {
	Status   string `json:"status"`
	Filename string `json:"filename"`
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}()

	if err := os.MkdirAll("uploads", 0755); err != nil {
		http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll("transcriptions", 0755); err != nil {
		http.Error(w, "Failed to create transcriptions directory", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join("uploads", handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := dst.Close(); err != nil {
			fmt.Printf("Error closing destination file: %v\n", err)
		}
	}()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to copy file", http.StatusInternalServerError)
		return
	}

	// Background processing
	go func(filename string) {
		cmd := exec.Command("/home/caleb/dev/python/qc_ai/venv/bin/python3", "/home/caleb/dev/python/qc_ai/main.py", filename, "../transcriptions/"+filename+".txt")
		cmd.Dir = "./uploads"

		tranJob := &jobqueue.Job{
			Type:        jobqueue.Transcription,
			Description: filename,
			Run: func() error {
				if out, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf("Error processing %s: %v\nOutput:\n%s\n", filename, err, out)
				} else {
					fmt.Printf("Finished processing %s\n", filename)
				}

				reportInput := filepath.Join("transcriptions", filename+".txt")
				reportOutput := filepath.Join("uploads", filename+".txt")
				genJob := &jobqueue.Job{
					Type:        jobqueue.ReportGeneration,
					Description: filename,
					Run: func() error {
						ReportGenerator.GenerateReport(reportInput, reportOutput)
						return nil
					},
				}
				jobqueue.Enqueue(genJob)
				return err
			},
		}
		jobqueue.Enqueue(tranJob)
	}(handler.Filename)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(UploadResponse{
		Status:   "uploaded",
		Filename: handler.Filename,
	}); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func UploadsHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir("./uploads")
	if err != nil {
		http.Error(w, "Error reading directory: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		fmt.Printf("Error encoding files response: %v\n", err)
	}
}
