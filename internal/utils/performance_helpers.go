package utils

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PerformanceReport struct {
	TestName        string        `json:"test_name"`
	Duration        time.Duration `json:"duration"`
	TotalRequests   int           `json:"total_requests"`
	SuccessfulReqs  int           `json:"successful_requests"`
	FailedRequests  int           `json:"failed_requests"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	P95ResponseTime time.Duration `json:"p95_response_time"`
	ThroughputRPS   float64       `json:"throughput_rps"`
	MaxMemoryMB     float64       `json:"max_memory_mb"`
	MaxGoroutines   int           `json:"max_goroutines"`
	Errors          []string      `json:"errors,omitempty"`
}

type PerformanceTracker struct {
	testName      string
	startTime     time.Time
	responses     []time.Duration
	errors        []error
	totalRequests int
	mu            sync.Mutex
	memStats      runtime.MemStats
	maxGoroutines int
}

func NewPerformanceTracker(testName string) *PerformanceTracker {
	return &PerformanceTracker{
		testName:  testName,
		startTime: time.Now(),
		responses: make([]time.Duration, 0),
		errors:    make([]error, 0),
	}
}

func (p *PerformanceTracker) RecordRequest(duration time.Duration, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalRequests++
	p.responses = append(p.responses, duration)
	if err != nil {
		p.errors = append(p.errors, err)
	}

	// Track goroutines and memory
	numGoroutines := runtime.NumGoroutine()
	if numGoroutines > p.maxGoroutines {
		p.maxGoroutines = numGoroutines
	}

	runtime.ReadMemStats(&p.memStats)
}

func (p *PerformanceTracker) GenerateReport() PerformanceReport {
	p.mu.Lock()
	defer p.mu.Unlock()

	duration := time.Since(p.startTime)
	successfulReqs := p.totalRequests - len(p.errors)

	// Calculate average response time
	var totalResponseTime time.Duration
	for _, resp := range p.responses {
		totalResponseTime += resp
	}
	avgResponseTime := time.Duration(0)
	if len(p.responses) > 0 {
		avgResponseTime = totalResponseTime / time.Duration(len(p.responses))
	}

	// Calculate P95 response time
	p95ResponseTime := time.Duration(0)
	if len(p.responses) > 0 {
		sortedResponses := make([]time.Duration, len(p.responses))
		copy(sortedResponses, p.responses)

		// Simple sort for P95 calculation
		for i := 0; i < len(sortedResponses); i++ {
			for j := i + 1; j < len(sortedResponses); j++ {
				if sortedResponses[i] > sortedResponses[j] {
					sortedResponses[i], sortedResponses[j] = sortedResponses[j], sortedResponses[i]
				}
			}
		}

		p95Index := int(float64(len(sortedResponses)) * 0.95)
		if p95Index >= len(sortedResponses) {
			p95Index = len(sortedResponses) - 1
		}
		p95ResponseTime = sortedResponses[p95Index]
	}

	// Calculate throughput
	throughputRPS := float64(successfulReqs) / duration.Seconds()

	// Convert memory to MB
	maxMemoryMB := float64(p.memStats.Alloc) / 1024 / 1024

	// Collect error messages
	errorMessages := make([]string, len(p.errors))
	for i, err := range p.errors {
		errorMessages[i] = err.Error()
	}

	return PerformanceReport{
		TestName:        p.testName,
		Duration:        duration,
		TotalRequests:   p.totalRequests,
		SuccessfulReqs:  successfulReqs,
		FailedRequests:  len(p.errors),
		AvgResponseTime: avgResponseTime,
		P95ResponseTime: p95ResponseTime,
		ThroughputRPS:   throughputRPS,
		MaxMemoryMB:     maxMemoryMB,
		MaxGoroutines:   p.maxGoroutines,
		Errors:          errorMessages,
	}
}

func SetupPerformanceTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect performance test database: " + err.Error())
	}

	// Configure connection pool for performance testing
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get underlying sql.DB: " + err.Error())
	}

	// Set connection pool settings for performance testing
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(10)                  // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum connection lifetime
	sqlDB.SetConnMaxIdleTime(time.Minute * 30) // Maximum connection idle time

	return db
}

func ExecuteConcurrentTest(name string, workers int, operationsPerWorker int, operation func(workerID int, operationID int) error) PerformanceReport {
	tracker := NewPerformanceTracker(name)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				start := time.Now()
				err := operation(workerID, j)
				duration := time.Since(start)

				tracker.RecordRequest(duration, err)
			}
		}(i)
	}

	wg.Wait()
	return tracker.GenerateReport()
}

type DatabaseStats struct {
	OpenConnections int           `json:"open_connections"`
	InUseConns      int           `json:"in_use_connections"`
	IdleConns       int           `json:"idle_connections"`
	WaitCount       int64         `json:"wait_count"`
	WaitDuration    time.Duration `json:"wait_duration"`
	MaxIdleClosed   int64         `json:"max_idle_closed"`
	MaxLifeClosed   int64         `json:"max_lifetime_closed"`
}

func GetDatabaseStats(db *gorm.DB) (DatabaseStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return DatabaseStats{}, err
	}

	stats := sqlDB.Stats()

	return DatabaseStats{
		OpenConnections: stats.OpenConnections,
		InUseConns:      stats.InUse,
		IdleConns:       stats.Idle,
		WaitCount:       stats.WaitCount,
		WaitDuration:    stats.WaitDuration,
		MaxIdleClosed:   stats.MaxIdleClosed,
		MaxLifeClosed:   stats.MaxLifetimeClosed,
	}, nil
}

func PrintDatabaseStats(db *gorm.DB, label string) {
	stats, err := GetDatabaseStats(db)
	if err != nil {
		fmt.Printf("Error getting DB stats for %s: %v\n", label, err)
		return
	}

	fmt.Printf("=== Database Stats (%s) ===\n", label)
	fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
	fmt.Printf("In Use: %d\n", stats.InUseConns)
	fmt.Printf("Idle: %d\n", stats.IdleConns)
	fmt.Printf("Wait Count: %d\n", stats.WaitCount)
	fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)
	fmt.Printf("Max Idle Closed: %d\n", stats.MaxIdleClosed)
	fmt.Printf("Max Lifetime Closed: %d\n", stats.MaxLifeClosed)
	fmt.Printf("==============================\n")
}

func StressTestDatabase(db *gorm.DB, duration time.Duration, maxConcurrent int) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	stop := make(chan bool)
	var wg sync.WaitGroup

	// Start the stress test
	go func() {
		time.Sleep(duration)
		close(stop)
	}()

	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				default:
					// Execute a simple query to stress the connection pool
					var result int
					err := sqlDB.QueryRow("SELECT 1").Scan(&result)
					if err != nil {
						fmt.Printf("Worker %d query error: %v\n", workerID, err)
					}

					// Small delay to prevent overwhelming
					time.Sleep(time.Millisecond * 10)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

func WaitForDB(db *gorm.DB, maxRetries int, retryDelay time.Duration) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	for i := 0; i < maxRetries; i++ {
		if err := sqlDB.Ping(); err == nil {
			return nil
		}
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("database not ready after %d retries", maxRetries)
}

func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func ForceCloseAllConnections(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Set connection limits to 0 to force close all connections
	sqlDB.SetMaxOpenConns(0)
	sqlDB.SetMaxIdleConns(0)

	// Give it a moment to close connections
	time.Sleep(time.Millisecond * 100)

	// Restore some minimal connection settings
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	return nil
}
