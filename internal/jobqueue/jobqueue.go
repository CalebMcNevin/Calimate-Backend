package jobqueue

import (
	"fmt"
	"time"
)

type JobType string

const (
	Transcription    JobType = "Transcription"
	ReportGeneration JobType = "ReportGeneration"
)

type Job struct {
	Type        JobType
	Description string
	// Cmd         *exec.Cmd
	Run func() error
}

// Channels for each job type
var (
	transcriptionChan    = make(chan *Job, 100)
	reportGenerationChan = make(chan *Job, 100)
)

func Enqueue(job *Job) {
	switch job.Type {
	case Transcription:
		transcriptionChan <- job
	case ReportGeneration:
		reportGenerationChan <- job
	default:
		fmt.Println("Unknown job type")
	}
}

func Worker() {
	var lastType JobType

	for {
		var job *Job

		// Try to get preferred job type first
		select {
		case job = <-preferChan(lastType):
			// Got preferred type
		default:
			// Preferred type not available; try any
			select {
			case job = <-transcriptionChan:
			case job = <-reportGenerationChan:
			default:
				// No jobs; wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		if job == nil {
			continue
		}

		fmt.Printf("[RUNNING] Job: %s (%s)\n", job.Description, job.Type)
		err := job.Run()
		if err != nil {
			fmt.Printf("[ERROR] %v\n", err)
		} else {
			fmt.Println("[DONE]")
		}
		lastType = job.Type
	}
}

func preferChan(last JobType) chan *Job {
	if last == Transcription {
		return transcriptionChan
	}
	return reportGenerationChan
}
