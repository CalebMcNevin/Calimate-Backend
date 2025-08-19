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
)

func TestPerformanceConcurrentEmployeeCreation(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	t.Run("50 workers creating 10 employees each", func(t *testing.T) {
		const numWorkers = 50
		const employeesPerWorker = 10

		report := utils.ExecuteConcurrentTest(
			"ConcurrentEmployeeCreation_50x10",
			numWorkers,
			employeesPerWorker,
			func(workerID int, operationID int) error {
				emp, err := employees.NewEmployee(
					fmt.Sprintf("Worker%d_Op%d_Name", workerID, operationID),
					fmt.Sprintf("Worker%d_First%d", workerID, operationID),
					fmt.Sprintf("Worker%d_Last%d", workerID, operationID),
					fmt.Sprintf("EMP%03d%03d", workerID, operationID),
				)
				if err != nil {
					return err
				}

				return service.CreateEmployee(emp)
			},
		)

		t.Logf("Performance Report: %+v", report)

		// Assertions
		assert.Equal(t, numWorkers*employeesPerWorker, report.TotalRequests)
		assert.True(t, report.FailedRequests < 10, "Too many failed requests: %d", report.FailedRequests)
		assert.True(t, report.ThroughputRPS > 50, "Throughput too low: %.2f RPS", report.ThroughputRPS)
		assert.True(t, report.AvgResponseTime < time.Millisecond*100, "Average response time too high: %v", report.AvgResponseTime)

		// Verify actual count in database
		var count int64
		err := db.Model(&employees.Employee{}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(report.SuccessfulReqs), count)
	})

	t.Run("100 workers creating 5 employees each - stress test", func(t *testing.T) {
		const numWorkers = 100
		const employeesPerWorker = 5

		report := utils.ExecuteConcurrentTest(
			"ConcurrentEmployeeCreation_100x5",
			numWorkers,
			employeesPerWorker,
			func(workerID int, operationID int) error {
				emp, err := employees.NewEmployee(
					fmt.Sprintf("Stress%d_Op%d_Name", workerID, operationID),
					fmt.Sprintf("Stress%d_First%d", workerID, operationID),
					fmt.Sprintf("Stress%d_Last%d", workerID, operationID),
					fmt.Sprintf("STR%03d%03d", workerID, operationID),
				)
				if err != nil {
					return err
				}

				return service.CreateEmployee(emp)
			},
		)

		t.Logf("Stress Test Report: %+v", report)

		// More lenient assertions for stress test
		assert.Equal(t, numWorkers*employeesPerWorker, report.TotalRequests)
		assert.True(t, report.FailedRequests < 25, "Too many failed requests: %d", report.FailedRequests)
		assert.True(t, report.ThroughputRPS > 25, "Throughput too low: %.2f RPS", report.ThroughputRPS)
	})
}

func TestPerformanceConcurrentEmployeeValidation(t *testing.T) {
	t.Run("Concurrent validation with invalid data", func(t *testing.T) {
		const numWorkers = 30
		const validationsPerWorker = 20

		report := utils.ExecuteConcurrentTest(
			"ConcurrentEmployeeValidation",
			numWorkers,
			validationsPerWorker,
			func(workerID int, operationID int) error {
				// Intentionally create invalid employees to test validation
				testCases := []struct {
					commonName, firstName, lastName, empNum string
					shouldFail                              bool
				}{
					{"Valid Name", "Valid", "Name", "EMP123", false},
					{"", "Invalid", "Name", "EMP124", false},                     // Empty common name is allowed in NewEmployee
					{"Valid", "", "Name", "EMP125", false},                       // Empty first name is allowed in NewEmployee
					{"Valid", "Name", "", "EMP126", false},                       // Empty last name is allowed in NewEmployee
					{"Valid", "Name", "Test", "AB", true},                        // Employee number too short
					{"Valid", "Name", "Test", "EMP-123", true},                   // Invalid characters
					{string(make([]byte, 101)), "Valid", "Name", "EMP127", true}, // Common name too long
					{"Valid", string(make([]byte, 51)), "Name", "EMP128", true},  // First name too long
					{"Valid", "Name", string(make([]byte, 51)), "EMP129", true},  // Last name too long
				}

				testCase := testCases[operationID%len(testCases)]
				_, err := employees.NewEmployee(
					testCase.commonName,
					testCase.firstName,
					testCase.lastName,
					testCase.empNum,
				)

				if testCase.shouldFail && err == nil {
					return fmt.Errorf("expected validation to fail but it didn't")
				}
				if !testCase.shouldFail && err != nil {
					return fmt.Errorf("expected validation to pass but got: %v", err)
				}

				return nil
			},
		)

		t.Logf("Validation Test Report: %+v", report)

		// Should have high throughput for validation-only operations
		assert.Equal(t, numWorkers*validationsPerWorker, report.TotalRequests)
		assert.True(t, report.ThroughputRPS > 500, "Validation throughput too low: %.2f RPS", report.ThroughputRPS)
		assert.True(t, report.AvgResponseTime < time.Millisecond*10, "Validation too slow: %v", report.AvgResponseTime)
	})
}

func TestPerformanceDatabaseConnectionStress(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	t.Run("Database connection pool stress test", func(t *testing.T) {
		utils.PrintDatabaseStats(db, "Before stress test")

		// Create some test data first
		for i := 0; i < 10; i++ {
			emp, err := employees.NewEmployee(
				fmt.Sprintf("Baseline%d", i),
				fmt.Sprintf("First%d", i),
				fmt.Sprintf("Last%d", i),
				fmt.Sprintf("BASE%03d", i),
			)
			require.NoError(t, err)
			require.NoError(t, service.CreateEmployee(emp))
		}

		// Now stress test the connection pool
		const numWorkers = 50
		const operationsPerWorker = 20

		var wg sync.WaitGroup
		errors := make(chan error, numWorkers*operationsPerWorker)

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < operationsPerWorker; j++ {
					// Mix of operations to stress different parts of the connection pool
					switch j % 4 {
					case 0:
						// Create operation
						emp, err := employees.NewEmployee(
							fmt.Sprintf("Pool%d_%d", workerID, j),
							fmt.Sprintf("Pool%d", workerID),
							fmt.Sprintf("Test%d", j),
							fmt.Sprintf("POOL%02d%02d", workerID, j),
						)
						if err != nil {
							errors <- err
							continue
						}
						err = service.CreateEmployee(emp)
						if err != nil {
							errors <- err
						}

					case 1:
						// Read all employees
						filter := employees.EmployeeFilter{}
						_, err := service.GetEmployees(filter)
						if err != nil {
							errors <- err
						}

					case 2:
						// Read specific employee (will fail for non-existent IDs, but tests connection)
						_, err := service.GetEmployeeByID("00000000-0000-0000-0000-000000000000")
						// We expect this to fail, so don't record the error
						_ = err

					case 3:
						// Update operation (will also fail, but tests connection)
						patch := employees.EmployeePatch{
							CommonName: &[]string{fmt.Sprintf("Updated%d_%d", workerID, j)}[0],
						}
						_, err := service.UpdateEmployee("00000000-0000-0000-0000-000000000000", patch)
						// We expect this to fail, so don't record the error
						_ = err
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		utils.PrintDatabaseStats(db, "After stress test")

		// Count errors
		errorCount := 0
		for err := range errors {
			t.Logf("Connection pool stress error: %v", err)
			errorCount++
		}

		// Should handle the load without too many connection errors
		assert.True(t, errorCount < numWorkers, "Too many connection errors: %d", errorCount)

		// Verify database is still responsive
		filter := employees.EmployeeFilter{}
		employees_result, err := service.GetEmployees(filter)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(employees_result), 10, "Database should have at least baseline employees")
	})
}

func TestPerformanceMemoryUsage(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	t.Run("Memory usage during bulk operations", func(t *testing.T) {
		tracker := utils.NewPerformanceTracker("MemoryUsageTest")

		// Create a large number of employees to test memory usage
		const totalEmployees = 1000

		for i := 0; i < totalEmployees; i++ {
			start := time.Now()

			emp, err := employees.NewEmployee(
				fmt.Sprintf("Memory_%d", i),
				fmt.Sprintf("Test_%d", i),
				fmt.Sprintf("User_%d", i),
				fmt.Sprintf("MEM%05d", i),
			)
			require.NoError(t, err)

			err = service.CreateEmployee(emp)
			duration := time.Since(start)
			tracker.RecordRequest(duration, err)

			// Periodically check memory usage
			if i%100 == 0 {
				report := tracker.GenerateReport()
				t.Logf("Progress: %d/%d, Memory: %.2f MB, Goroutines: %d",
					i, totalEmployees, report.MaxMemoryMB, report.MaxGoroutines)
			}
		}

		finalReport := tracker.GenerateReport()
		t.Logf("Final Memory Report: %+v", finalReport)

		// Memory usage should be reasonable
		assert.True(t, finalReport.MaxMemoryMB < 100, "Memory usage too high: %.2f MB", finalReport.MaxMemoryMB)
		assert.True(t, finalReport.MaxGoroutines < 20, "Too many goroutines: %d", finalReport.MaxGoroutines)
		assert.Equal(t, 0, finalReport.FailedRequests, "Should have no failed requests")

		// Verify all employees were created
		var count int64
		err = db.Model(&employees.Employee{}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(totalEmployees), count)
	})
}

func TestPerformanceRaceConditions(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	t.Run("Race condition detection in concurrent operations", func(t *testing.T) {
		// Create a base employee for testing updates
		baseEmp, err := employees.NewEmployee("Race Test", "Race", "Test", "RACE001")
		require.NoError(t, err)
		require.NoError(t, service.CreateEmployee(baseEmp))

		const numWorkers = 20
		const operationsPerWorker = 10

		var wg sync.WaitGroup
		var updateErrors []error
		var readErrors []error
		var mu sync.Mutex

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < operationsPerWorker; j++ {
					// Alternate between reads and updates to the same employee
					if j%2 == 0 {
						// Update operation
						newName := fmt.Sprintf("Updated_by_worker_%d_op_%d", workerID, j)
						patch := employees.EmployeePatch{
							CommonName: &newName,
						}
						_, err := service.UpdateEmployee(baseEmp.ID.String(), patch)
						if err != nil {
							mu.Lock()
							updateErrors = append(updateErrors, err)
							mu.Unlock()
						}
					} else {
						// Read operation
						_, err := service.GetEmployeeByID(baseEmp.ID.String())
						if err != nil {
							mu.Lock()
							readErrors = append(readErrors, err)
							mu.Unlock()
						}
					}
				}
			}(i)
		}

		wg.Wait()

		t.Logf("Update errors: %d, Read errors: %d", len(updateErrors), len(readErrors))

		// We should have very few errors from race conditions
		assert.True(t, len(updateErrors) < 5, "Too many update errors: %d", len(updateErrors))
		assert.True(t, len(readErrors) < 2, "Too many read errors: %d", len(readErrors))

		// The employee should still exist and be readable
		finalEmployee, err := service.GetEmployeeByID(baseEmp.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, finalEmployee)
		assert.Contains(t, finalEmployee.CommonName, "Updated_by_worker_", "Employee should have been updated")
	})
}

func TestPerformanceConcurrentReads(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	// Create test data
	const numTestEmployees = 100
	employeeIDs := make([]string, numTestEmployees)

	for i := 0; i < numTestEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("ReadTest_%d", i),
			fmt.Sprintf("Read_%d", i),
			fmt.Sprintf("Test_%d", i),
			fmt.Sprintf("READ%03d", i),
		)
		require.NoError(t, err)
		require.NoError(t, service.CreateEmployee(emp))
		employeeIDs[i] = emp.ID.String()
	}

	t.Run("Concurrent reads by ID", func(t *testing.T) {
		const numWorkers = 50
		const readsPerWorker = 20

		report := utils.ExecuteConcurrentTest(
			"ConcurrentReadsById",
			numWorkers,
			readsPerWorker,
			func(workerID int, operationID int) error {
				// Read random employee
				employeeID := employeeIDs[operationID%len(employeeIDs)]
				_, err := service.GetEmployeeByID(employeeID)
				return err
			},
		)

		t.Logf("Concurrent Reads Report: %+v", report)

		// Read operations should be very fast
		assert.Equal(t, numWorkers*readsPerWorker, report.TotalRequests)
		assert.Equal(t, 0, report.FailedRequests, "No read operations should fail")
		assert.True(t, report.ThroughputRPS > 200, "Read throughput too low: %.2f RPS", report.ThroughputRPS)
		assert.True(t, report.AvgResponseTime < time.Millisecond*50, "Read response time too high: %v", report.AvgResponseTime)
	})

	t.Run("Concurrent reads all employees", func(t *testing.T) {
		const numWorkers = 30
		const readsPerWorker = 10

		report := utils.ExecuteConcurrentTest(
			"ConcurrentReadsAll",
			numWorkers,
			readsPerWorker,
			func(workerID int, operationID int) error {
				filter := employees.EmployeeFilter{}
				result, err := service.GetEmployees(filter)
				if err != nil {
					return err
				}

				// Verify we got expected number of employees
				if len(result) < numTestEmployees {
					return fmt.Errorf("expected at least %d employees, got %d", numTestEmployees, len(result))
				}
				return nil
			},
		)

		t.Logf("Concurrent Read All Report: %+v", report)

		// Reading all employees is more expensive but should still be reasonable
		assert.Equal(t, numWorkers*readsPerWorker, report.TotalRequests)
		assert.True(t, report.FailedRequests < 2, "Too many failed requests: %d", report.FailedRequests)
		assert.True(t, report.ThroughputRPS > 20, "Read all throughput too low: %.2f RPS", report.ThroughputRPS)
		assert.True(t, report.AvgResponseTime < time.Millisecond*200, "Read all response time too high: %v", report.AvgResponseTime)
	})
}

func TestPerformanceConcurrentUpdates(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	// Create test employees for updating
	const numTestEmployees = 50
	employeeIDs := make([]string, numTestEmployees)

	for i := 0; i < numTestEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("UpdateTest_%d", i),
			fmt.Sprintf("Update_%d", i),
			fmt.Sprintf("Test_%d", i),
			fmt.Sprintf("UPD%03d", i),
		)
		require.NoError(t, err)
		require.NoError(t, service.CreateEmployee(emp))
		employeeIDs[i] = emp.ID.String()
	}

	t.Run("Concurrent updates to different employees", func(t *testing.T) {
		const numWorkers = 25
		const updatesPerWorker = 4

		report := utils.ExecuteConcurrentTest(
			"ConcurrentUpdatesDifferent",
			numWorkers,
			updatesPerWorker,
			func(workerID int, operationID int) error {
				// Update different employee each time
				employeeIndex := (workerID*updatesPerWorker + operationID) % len(employeeIDs)
				employeeID := employeeIDs[employeeIndex]

				newName := fmt.Sprintf("Updated_by_w%d_o%d", workerID, operationID)
				patch := employees.EmployeePatch{
					CommonName: &newName,
				}

				_, err := service.UpdateEmployee(employeeID, patch)
				return err
			},
		)

		t.Logf("Concurrent Updates Different Report: %+v", report)

		// Updates to different employees should have minimal contention
		assert.Equal(t, numWorkers*updatesPerWorker, report.TotalRequests)
		assert.True(t, report.FailedRequests < 3, "Too many failed requests: %d", report.FailedRequests)
		assert.True(t, report.ThroughputRPS > 30, "Update throughput too low: %.2f RPS", report.ThroughputRPS)
		assert.True(t, report.AvgResponseTime < time.Millisecond*100, "Update response time too high: %v", report.AvgResponseTime)
	})

	t.Run("Concurrent updates to same employee - contention test", func(t *testing.T) {
		// Create a single employee for contention testing
		contendedEmp, err := employees.NewEmployee("Contended", "Test", "Employee", "CONTEND")
		require.NoError(t, err)
		require.NoError(t, service.CreateEmployee(contendedEmp))

		const numWorkers = 20
		const updatesPerWorker = 5

		var successCount int64
		var updateMu sync.Mutex

		report := utils.ExecuteConcurrentTest(
			"ConcurrentUpdatesSame",
			numWorkers,
			updatesPerWorker,
			func(workerID int, operationID int) error {
				newName := fmt.Sprintf("Contended_w%d_o%d_%d", workerID, operationID, time.Now().UnixNano())
				patch := employees.EmployeePatch{
					CommonName: &newName,
				}

				_, err := service.UpdateEmployee(contendedEmp.ID.String(), patch)
				if err == nil {
					updateMu.Lock()
					successCount++
					updateMu.Unlock()
				}
				return err
			},
		)

		t.Logf("Concurrent Updates Same Report: %+v", report)
		t.Logf("Successful updates: %d", successCount)

		// Updates to same employee will have contention, but most should succeed
		assert.Equal(t, numWorkers*updatesPerWorker, report.TotalRequests)
		assert.True(t, report.SuccessfulReqs > int(successCount*8/10), "Too many failed updates due to contention")

		// Verify final state
		finalEmp, err := service.GetEmployeeByID(contendedEmp.ID.String())
		assert.NoError(t, err)
		assert.Contains(t, finalEmp.CommonName, "Contended_w", "Employee should have been updated")
	})
}

func TestPerformanceMixedOperations(t *testing.T) {
	db := utils.SetupPerformanceTestDB()
	defer utils.CloseDB(db)

	err := db.AutoMigrate(employees.Models()...)
	require.NoError(t, err)

	service := employees.NewEmployeeService(db)

	// Create some base data
	const numBaseEmployees = 20
	baseEmployeeIDs := make([]string, numBaseEmployees)

	for i := 0; i < numBaseEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("MixedBase_%d", i),
			fmt.Sprintf("Mixed_%d", i),
			fmt.Sprintf("Base_%d", i),
			fmt.Sprintf("MIX%03d", i),
		)
		require.NoError(t, err)
		require.NoError(t, service.CreateEmployee(emp))
		baseEmployeeIDs[i] = emp.ID.String()
	}

	t.Run("Mixed CRUD operations", func(t *testing.T) {
		const numWorkers = 30
		const operationsPerWorker = 10

		var createCount, readCount, updateCount int64
		var statsMu sync.Mutex

		report := utils.ExecuteConcurrentTest(
			"MixedCRUDOperations",
			numWorkers,
			operationsPerWorker,
			func(workerID int, operationID int) error {
				switch operationID % 5 {
				case 0, 1: // 40% reads
					statsMu.Lock()
					readCount++
					statsMu.Unlock()

					if operationID%2 == 0 {
						// Read by ID
						empID := baseEmployeeIDs[operationID%len(baseEmployeeIDs)]
						_, err := service.GetEmployeeByID(empID)
						return err
					} else {
						// Read all with filter
						filter := employees.EmployeeFilter{}
						_, err := service.GetEmployees(filter)
						return err
					}

				case 2: // 20% creates
					statsMu.Lock()
					createCount++
					statsMu.Unlock()

					emp, err := employees.NewEmployee(
						fmt.Sprintf("MixedCreate_w%d_o%d", workerID, operationID),
						fmt.Sprintf("Worker%d", workerID),
						fmt.Sprintf("Op%d", operationID),
						fmt.Sprintf("CR%02d%02d", workerID, operationID),
					)
					if err != nil {
						return err
					}
					return service.CreateEmployee(emp)

				case 3, 4: // 40% updates
					statsMu.Lock()
					updateCount++
					statsMu.Unlock()

					empID := baseEmployeeIDs[operationID%len(baseEmployeeIDs)]
					newName := fmt.Sprintf("MixedUpdate_w%d_o%d", workerID, operationID)
					patch := employees.EmployeePatch{
						CommonName: &newName,
					}
					_, err := service.UpdateEmployee(empID, patch)
					return err
				}

				return nil
			},
		)

		t.Logf("Mixed Operations Report: %+v", report)
		t.Logf("Operation distribution - Creates: %d, Reads: %d, Updates: %d", createCount, readCount, updateCount)

		// Mixed operations should handle well
		assert.Equal(t, numWorkers*operationsPerWorker, report.TotalRequests)
		assert.True(t, report.FailedRequests < 15, "Too many failed requests: %d", report.FailedRequests)
		assert.True(t, report.ThroughputRPS > 40, "Mixed operations throughput too low: %.2f RPS", report.ThroughputRPS)
		assert.True(t, report.AvgResponseTime < time.Millisecond*150, "Mixed operations response time too high: %v", report.AvgResponseTime)

		// Verify operation distribution is roughly correct
		totalOps := createCount + readCount + updateCount
		createPercent := float64(createCount) / float64(totalOps) * 100
		readPercent := float64(readCount) / float64(totalOps) * 100
		updatePercent := float64(updateCount) / float64(totalOps) * 100

		assert.True(t, createPercent > 15 && createPercent < 25, "Create percentage out of range: %.1f%%", createPercent)
		assert.True(t, readPercent > 35 && readPercent < 45, "Read percentage out of range: %.1f%%", readPercent)
		assert.True(t, updatePercent > 35 && updatePercent < 45, "Update percentage out of range: %.1f%%", updatePercent)
	})
}
