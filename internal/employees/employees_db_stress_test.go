package employees_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"qc_api/internal/employees"
	"qc_api/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseConnectionPoolStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database stress tests in short mode")
	}

	t.Run("Connection pool exhaustion test", func(t *testing.T) {
		// Create a database with limited connection pool
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		require.NoError(t, err)
		defer utils.CloseDB(db)

		// Set very limited connection pool to force contention
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.SetMaxOpenConns(5) // Very limited
		sqlDB.SetMaxIdleConns(2) // Even more limited
		sqlDB.SetConnMaxLifetime(time.Minute)
		sqlDB.SetConnMaxIdleTime(time.Second * 30)

		err = db.AutoMigrate(employees.Models()...)
		require.NoError(t, err)

		service := employees.NewEmployeeService(db)

		utils.PrintDatabaseStats(db, "Before connection pool stress")

		const numWorkers = 20 // More workers than connections
		const operationsPerWorker = 15

		var wg sync.WaitGroup
		errors := make(chan error, numWorkers*operationsPerWorker)
		successes := make(chan bool, numWorkers*operationsPerWorker)

		startTime := time.Now()

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < operationsPerWorker; j++ {
					// Mix of operations to stress connection pool
					switch j % 3 {
					case 0:
						// Create operation
						emp, err := employees.NewEmployee(
							fmt.Sprintf("ConnPool_w%d_o%d", workerID, j),
							fmt.Sprintf("Worker%d", workerID),
							fmt.Sprintf("Op%d", j),
							fmt.Sprintf("CP%02d%02d", workerID, j),
						)
						if err != nil {
							errors <- err
							continue
						}

						err = service.CreateEmployee(emp)
						if err != nil {
							errors <- err
						} else {
							successes <- true
						}

					case 1:
						// Read all employees
						filter := employees.EmployeeFilter{}
						_, err := service.GetEmployees(filter)
						if err != nil {
							errors <- err
						} else {
							successes <- true
						}

					case 2:
						// Simple database ping to test connection availability
						sqlDB, err := db.DB()
						if err != nil {
							errors <- err
							continue
						}

						err = sqlDB.Ping()
						if err != nil {
							errors <- err
						} else {
							successes <- true
						}
					}

					// Add small delay to allow connection cycling
					time.Sleep(time.Millisecond * 5)
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		close(successes)

		duration := time.Since(startTime)

		utils.PrintDatabaseStats(db, "After connection pool stress")

		// Count results
		errorCount := len(errors)
		successCount := len(successes)
		totalOps := numWorkers * operationsPerWorker

		t.Logf("Connection Pool Stress Results:")
		t.Logf("Duration: %v", duration)
		t.Logf("Total Operations: %d", totalOps)
		t.Logf("Successful: %d", successCount)
		t.Logf("Failed: %d", errorCount)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(totalOps)*100)

		// Log some errors for debugging
		errorList := make([]error, 0, errorCount)
		for err := range errors {
			errorList = append(errorList, err)
		}
		for i, err := range errorList {
			if i < 5 { // Log first 5 errors
				t.Logf("Error %d: %v", i+1, err)
			}
		}

		// Even with limited connections, most operations should succeed
		// The connection pool should handle the contention gracefully
		assert.True(t, successCount > totalOps*7/10, "Too many operations failed: %d/%d", errorCount, totalOps)
		assert.True(t, errorCount < totalOps*3/10, "Error rate too high: %d/%d", errorCount, totalOps)
	})

	t.Run("Connection leak detection", func(t *testing.T) {
		db := utils.SetupPerformanceTestDB()
		defer utils.CloseDB(db)

		err := db.AutoMigrate(employees.Models()...)
		require.NoError(t, err)

		service := employees.NewEmployeeService(db)

		// Get baseline stats
		initialStats, err := utils.GetDatabaseStats(db)
		require.NoError(t, err)

		t.Logf("Initial DB Stats: %+v", initialStats)

		// Perform many operations and check for connection leaks
		const numOperations = 200

		for i := 0; i < numOperations; i++ {
			switch i % 4 {
			case 0:
				// Create
				emp, err := employees.NewEmployee(
					fmt.Sprintf("LeakTest_%d", i),
					fmt.Sprintf("Leak_%d", i),
					fmt.Sprintf("Test_%d", i),
					fmt.Sprintf("LT%05d", i),
				)
				require.NoError(t, err)
				require.NoError(t, service.CreateEmployee(emp))

			case 1:
				// Read all
				filter := employees.EmployeeFilter{}
				_, err := service.GetEmployees(filter)
				require.NoError(t, err)

			case 2:
				// Try to read non-existent employee
				_, err := service.GetEmployeeByID("00000000-0000-0000-0000-000000000000")
				// This should fail, but shouldn't leak connections
				_ = err

			case 3:
				// Try to update non-existent employee
				patch := employees.EmployeePatch{
					CommonName: &[]string{fmt.Sprintf("NonExistent_%d", i)}[0],
				}
				_, err := service.UpdateEmployee("00000000-0000-0000-0000-000000000000", patch)
				// This should fail, but shouldn't leak connections
				_ = err
			}

			// Periodically check connection stats
			if i%50 == 0 && i > 0 {
				stats, err := utils.GetDatabaseStats(db)
				require.NoError(t, err)
				t.Logf("Stats after %d operations: %+v", i, stats)
			}
		}

		// Wait a moment for connections to be returned to pool
		time.Sleep(time.Millisecond * 100)

		// Get final stats
		finalStats, err := utils.GetDatabaseStats(db)
		require.NoError(t, err)

		t.Logf("Final DB Stats: %+v", finalStats)

		// Check for connection leaks
		// The number of open connections shouldn't grow significantly
		assert.True(t, finalStats.OpenConnections <= initialStats.OpenConnections+3,
			"Possible connection leak detected: initial=%d, final=%d",
			initialStats.OpenConnections, finalStats.OpenConnections)

		// In-use connections should be low after operations complete
		assert.True(t, finalStats.InUseConns <= 2,
			"Too many connections still in use: %d", finalStats.InUseConns)
	})

	t.Run("Database timeout and recovery", func(t *testing.T) {
		db := utils.SetupPerformanceTestDB()
		defer utils.CloseDB(db)

		// Set aggressive connection timeouts
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Second * 30) // Short lifetime
		sqlDB.SetConnMaxIdleTime(time.Second * 10) // Short idle time

		err = db.AutoMigrate(employees.Models()...)
		require.NoError(t, err)

		service := employees.NewEmployeeService(db)

		// Create some test data
		for i := 0; i < 5; i++ {
			emp, err := employees.NewEmployee(
				fmt.Sprintf("TimeoutTest_%d", i),
				fmt.Sprintf("Timeout_%d", i),
				fmt.Sprintf("Test_%d", i),
				fmt.Sprintf("TO%05d", i),
			)
			require.NoError(t, err)
			require.NoError(t, service.CreateEmployee(emp))
		}

		utils.PrintDatabaseStats(db, "Before timeout test")

		// Wait for connections to timeout
		t.Log("Waiting for connections to timeout...")
		time.Sleep(time.Second * 12) // Wait longer than idle timeout

		utils.PrintDatabaseStats(db, "After timeout period")

		// Try operations after timeout - should still work
		t.Log("Testing operations after timeout...")

		// These operations should succeed and create new connections
		filter := employees.EmployeeFilter{}
		employees_result, err := service.GetEmployees(filter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(employees_result), 5)

		// Create new employee
		newEmp, err := employees.NewEmployee("PostTimeout", "Post", "Timeout", "POST01")
		require.NoError(t, err)
		assert.NoError(t, service.CreateEmployee(newEmp))

		utils.PrintDatabaseStats(db, "After recovery operations")

		t.Log("Database recovered successfully from connection timeout")
	})

	t.Run("High frequency operations stress", func(t *testing.T) {
		db := utils.SetupPerformanceTestDB()
		defer utils.CloseDB(db)

		err := db.AutoMigrate(employees.Models()...)
		require.NoError(t, err)

		service := employees.NewEmployeeService(db)

		// High frequency, short duration test
		const duration = time.Second * 5
		const maxConcurrent = 15

		var operationCount int64
		var errorCount int64
		var mu sync.Mutex

		stop := make(chan bool)
		var wg sync.WaitGroup

		// Start the timer
		go func() {
			time.Sleep(duration)
			close(stop)
		}()

		startTime := time.Now()

		for i := 0; i < maxConcurrent; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				localOps := 0
				localErrors := 0

				for {
					select {
					case <-stop:
						mu.Lock()
						operationCount += int64(localOps)
						errorCount += int64(localErrors)
						mu.Unlock()
						return
					default:
						// Perform rapid operations
						if localOps%3 == 0 {
							// Create
							emp, err := employees.NewEmployee(
								fmt.Sprintf("HighFreq_w%d_o%d", workerID, localOps),
								fmt.Sprintf("Worker%d", workerID),
								fmt.Sprintf("Op%d", localOps),
								fmt.Sprintf("HF%02d%04d", workerID, localOps),
							)
							if err != nil {
								localErrors++
								continue
							}

							err = service.CreateEmployee(emp)
							if err != nil {
								localErrors++
							}
						} else {
							// Read
							filter := employees.EmployeeFilter{}
							_, err := service.GetEmployees(filter)
							if err != nil {
								localErrors++
							}
						}

						localOps++

						// Small delay to prevent overwhelming
						time.Sleep(time.Millisecond * 2)
					}
				}
			}(i)
		}

		wg.Wait()
		actualDuration := time.Since(startTime)

		utils.PrintDatabaseStats(db, "After high frequency stress")

		opsPerSecond := float64(operationCount) / actualDuration.Seconds()
		errorRate := float64(errorCount) / float64(operationCount) * 100

		t.Logf("High Frequency Stress Results:")
		t.Logf("Duration: %v", actualDuration)
		t.Logf("Total Operations: %d", operationCount)
		t.Logf("Operations/Second: %.2f", opsPerSecond)
		t.Logf("Errors: %d", errorCount)
		t.Logf("Error Rate: %.2f%%", errorRate)

		// High frequency operations should maintain reasonable performance
		assert.True(t, operationCount > 100, "Too few operations completed: %d", operationCount)
		assert.True(t, opsPerSecond > 20, "Operations per second too low: %.2f", opsPerSecond)
		assert.True(t, errorRate < 10, "Error rate too high: %.2f%%", errorRate)
	})
}
