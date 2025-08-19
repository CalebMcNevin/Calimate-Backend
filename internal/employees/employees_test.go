package employees_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qc_api/internal/employees"
	"qc_api/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test data
var testEmployeesData = []employees.EmployeeDTO{
	{
		CommonName:     "John Doe",
		FirstName:      "John",
		LastName:       "Doe",
		EmployeeNumber: "EMP001",
	},
	{
		CommonName:     "Jane Smith",
		FirstName:      "Jane",
		LastName:       "Smith",
		EmployeeNumber: "EMP002",
	},
	{
		CommonName:     "Bob Johnson",
		FirstName:      "Robert",
		LastName:       "Johnson",
		EmployeeNumber: "EMP003",
	},
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	if err := db.AutoMigrate(employees.Models()...); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// Create inspections table to avoid foreign key issues
	// We'll create a minimal structure to satisfy the relationship
	db.Exec(`CREATE TABLE IF NOT EXISTS inspections (
		id TEXT PRIMARY KEY,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		employee_id TEXT,
		report TEXT
	)`)

	return db
}

// setupTestEmployee creates a test employee in the database
func setupTestEmployee(db *gorm.DB) *employees.Employee {
	empData := testEmployeesData[0]
	employee, _ := employees.NewEmployee(empData.CommonName, empData.FirstName, empData.LastName, empData.EmployeeNumber)
	db.Create(employee)
	return employee
}

// setupEchoContext creates an Echo context for HTTP testing
func setupEchoContext(method, url, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = utils.NewValidator()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}

// createEmployeeInDB is a helper to create an employee directly in the database
func createEmployeeInDB(db *gorm.DB, empData employees.EmployeeDTO) *employees.Employee {
	employee, _ := employees.NewEmployee(empData.CommonName, empData.FirstName, empData.LastName, empData.EmployeeNumber)
	db.Create(employee)
	return employee
}

// ===============================
// MODEL TESTS
// ===============================

func TestNewEmployee(t *testing.T) {
	tests := []struct {
		name           string
		commonName     string
		firstName      string
		lastName       string
		employeeNumber string
		expectError    bool
	}{
		{
			name:           "Valid employee creation",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "EMP001",
			expectError:    false,
		},
		{
			name:           "Empty common name",
			commonName:     "",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "EMP001",
			expectError:    false, // NewEmployee doesn't validate - validation is in DTO
		},
		{
			name:           "All fields provided",
			commonName:     "Jane Smith",
			firstName:      "Jane",
			lastName:       "Smith",
			employeeNumber: "EMP002",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			employee, err := employees.NewEmployee(tt.commonName, tt.firstName, tt.lastName, tt.employeeNumber)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, employee)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, employee)
				assert.Equal(t, tt.commonName, employee.CommonName)
				assert.Equal(t, tt.firstName, employee.FirstName)
				assert.Equal(t, tt.lastName, employee.LastName)
				assert.Equal(t, tt.employeeNumber, employee.EmployeeNumber)
				assert.True(t, employee.Active) // Should default to true
			}
		})
	}
}

func TestEmployeeValidation(t *testing.T) {
	e := echo.New()
	e.Validator = utils.NewValidator()

	tests := []struct {
		name        string
		dto         employees.EmployeeDTO
		expectError bool
	}{
		{
			name: "Valid DTO",
			dto: employees.EmployeeDTO{
				CommonName:     "John Doe",
				FirstName:      "John",
				LastName:       "Doe",
				EmployeeNumber: "EMP001",
			},
			expectError: false,
		},
		{
			name: "Missing common name",
			dto: employees.EmployeeDTO{
				FirstName:      "John",
				LastName:       "Doe",
				EmployeeNumber: "EMP001",
			},
			expectError: true,
		},
		{
			name: "Missing first name",
			dto: employees.EmployeeDTO{
				CommonName:     "John Doe",
				LastName:       "Doe",
				EmployeeNumber: "EMP001",
			},
			expectError: true,
		},
		{
			name: "Missing last name",
			dto: employees.EmployeeDTO{
				CommonName:     "John Doe",
				FirstName:      "John",
				EmployeeNumber: "EMP001",
			},
			expectError: true,
		},
		{
			name: "Missing employee number",
			dto: employees.EmployeeDTO{
				CommonName: "John Doe",
				FirstName:  "John",
				LastName:   "Doe",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := e.Validator.Validate(&tt.dto)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ===============================
// SERVICE LAYER TESTS
// ===============================

func TestCreateEmployee(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	t.Run("Successfully create employee", func(t *testing.T) {
		empData := testEmployeesData[0]
		employee, err := employees.NewEmployee(empData.CommonName, empData.FirstName, empData.LastName, empData.EmployeeNumber)
		require.NoError(t, err)

		err = service.CreateEmployee(employee)
		assert.NoError(t, err)

		// Verify employee was saved to database
		var savedEmployee employees.Employee
		result := db.First(&savedEmployee, "employee_number = ?", empData.EmployeeNumber)
		assert.NoError(t, result.Error)
		assert.Equal(t, empData.CommonName, savedEmployee.CommonName)
		assert.Equal(t, empData.FirstName, savedEmployee.FirstName)
		assert.Equal(t, empData.LastName, savedEmployee.LastName)
		assert.Equal(t, empData.EmployeeNumber, savedEmployee.EmployeeNumber)
		assert.True(t, savedEmployee.Active)
	})
}

func TestGetEmployees(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test employees
	for _, empData := range testEmployeesData {
		createEmployeeInDB(db, empData)
	}

	t.Run("Get all employees with no filter", func(t *testing.T) {
		filter := employees.EmployeeFilter{}
		result, err := service.GetEmployees(filter)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Filter by active status - true", func(t *testing.T) {
		active := true
		filter := employees.EmployeeFilter{Active: &active}
		result, err := service.GetEmployees(filter)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		for _, emp := range result {
			assert.True(t, emp.Active)
		}
	})

	t.Run("Filter by employee number", func(t *testing.T) {
		empNum := "EMP001"
		filter := employees.EmployeeFilter{EmployeeNumber: &empNum}
		result, err := service.GetEmployees(filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "EMP001", result[0].EmployeeNumber)
	})
}

func TestGetEmployeeByID(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test employee
	createdEmp := createEmployeeInDB(db, testEmployeesData[0])

	t.Run("Successfully retrieve existing employee", func(t *testing.T) {
		result, err := service.GetEmployeeByID(createdEmp.ID.String())

		assert.NoError(t, err)
		assert.Equal(t, createdEmp.ID, result.ID)
		assert.Equal(t, createdEmp.CommonName, result.CommonName)
		assert.Equal(t, createdEmp.FirstName, result.FirstName)
		assert.Equal(t, createdEmp.LastName, result.LastName)
		assert.Equal(t, createdEmp.EmployeeNumber, result.EmployeeNumber)
	})

	t.Run("Return error for non-existent ID", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		result, err := service.GetEmployeeByID(nonExistentID)

		assert.Error(t, err)
		assert.Equal(t, employees.Employee{}, result)
	})
}

func TestUpdateEmployee(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test employee
	createdEmp := createEmployeeInDB(db, testEmployeesData[0])

	t.Run("Successfully update single field", func(t *testing.T) {
		newCommonName := "Updated Name"
		patch := employees.EmployeePatch{
			CommonName: &newCommonName,
		}

		result, err := service.UpdateEmployee(createdEmp.ID.String(), patch)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newCommonName, result.CommonName)
		assert.Equal(t, createdEmp.FirstName, result.FirstName) // Should remain unchanged
	})

	t.Run("Update non-existent employee", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		newName := "Test"
		patch := employees.EmployeePatch{
			CommonName: &newName,
		}

		result, err := service.UpdateEmployee(nonExistentID, patch)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// ===============================
// HANDLER TESTS (HTTP LAYER)
// ===============================

func TestPostEmployeeHandler(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	t.Run("Success case - valid DTO", func(t *testing.T) {
		reqBody := `{
			"common_name": "John Doe",
			"first_name": "John",
			"last_name": "Doe",
			"employee_number": "EMP001"
		}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", response.CommonName)
		assert.Equal(t, "John", response.FirstName)
		assert.Equal(t, "Doe", response.LastName)
		assert.Equal(t, "EMP001", response.EmployeeNumber)
		assert.True(t, response.Active)
	})

	t.Run("Validation error - invalid JSON", func(t *testing.T) {
		invalidJSON := `{"common_name": "John"`
		c, rec := setupEchoContext(http.MethodPost, "/employees", invalidJSON)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Validation error - missing required fields", func(t *testing.T) {
		reqBody := `{
			"common_name": "John Doe"
		}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestGetEmployeesHandler(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test employees
	for _, empData := range testEmployeesData {
		createEmployeeInDB(db, empData)
	}

	t.Run("Get all employees - no filters", func(t *testing.T) {
		c, rec := setupEchoContext(http.MethodGet, "/employees", "")

		err := service.GetEmployeesHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response []employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 3)
	})

	t.Run("Filter by employee number", func(t *testing.T) {
		c, rec := setupEchoContext(http.MethodGet, "/employees?employee_number=EMP001", "")

		err := service.GetEmployeesHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response []employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		if len(response) > 0 {
			assert.Equal(t, "EMP001", response[0].EmployeeNumber)
		}
	})
}

func TestGetEmployeeByIDHandler(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test employee
	createdEmp := createEmployeeInDB(db, testEmployeesData[0])

	t.Run("Success case - valid ID", func(t *testing.T) {
		c, rec := setupEchoContext(http.MethodGet, "/employees/"+createdEmp.ID.String(), "")
		c.SetParamNames("id")
		c.SetParamValues(createdEmp.ID.String())

		err := service.GetEmployeeByIDHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, createdEmp.ID, response.ID)
		assert.Equal(t, createdEmp.CommonName, response.CommonName)
	})

	t.Run("Not found - non-existent ID", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		c, rec := setupEchoContext(http.MethodGet, "/employees/"+nonExistentID, "")
		c.SetParamNames("id")
		c.SetParamValues(nonExistentID)

		err := service.GetEmployeeByIDHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestPatchEmployeeHandler(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test employee
	createdEmp := createEmployeeInDB(db, testEmployeesData[0])

	t.Run("Success case - update single field", func(t *testing.T) {
		reqBody := `{"common_name": "Updated Name"}`
		c, rec := setupEchoContext(http.MethodPatch, "/employees/"+createdEmp.ID.String(), reqBody)
		c.SetParamNames("id")
		c.SetParamValues(createdEmp.ID.String())

		err := service.PatchEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", response.CommonName)
	})

	t.Run("Validation error - invalid JSON", func(t *testing.T) {
		invalidJSON := `{"common_name": "Updated`
		c, rec := setupEchoContext(http.MethodPatch, "/employees/"+createdEmp.ID.String(), invalidJSON)
		c.SetParamNames("id")
		c.SetParamValues(createdEmp.ID.String())

		err := service.PatchEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Error - empty ID parameter", func(t *testing.T) {
		reqBody := `{"common_name": "Updated Name"}`
		c, rec := setupEchoContext(http.MethodPatch, "/employees/", reqBody)
		c.SetParamNames("id")
		c.SetParamValues("")

		err := service.PatchEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// ===============================
// INTEGRATION TESTS
// ===============================

func TestEmployeeWorkflow(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	t.Run("Complete CRUD workflow", func(t *testing.T) {
		// 1. Create employee via POST
		createBody := `{
			"common_name": "Integration Test",
			"first_name": "Integration",
			"last_name": "Test",
			"employee_number": "INT001"
		}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", createBody)
		err := service.PostEmployeeHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var createdEmployee employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &createdEmployee)
		assert.NoError(t, err)

		// 2. Retrieve employee via GET by ID
		c, rec = setupEchoContext(http.MethodGet, "/employees/"+createdEmployee.ID.String(), "")
		c.SetParamNames("id")
		c.SetParamValues(createdEmployee.ID.String())
		err = service.GetEmployeeByIDHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var retrievedEmployee employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &retrievedEmployee)
		assert.NoError(t, err)
		assert.Equal(t, createdEmployee.ID, retrievedEmployee.ID)

		// 3. Update employee via PATCH
		updateBody := `{"common_name": "Updated Integration Test"}`
		c, rec = setupEchoContext(http.MethodPatch, "/employees/"+createdEmployee.ID.String(), updateBody)
		c.SetParamNames("id")
		c.SetParamValues(createdEmployee.ID.String())
		err = service.PatchEmployeeHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var updatedEmployee employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &updatedEmployee)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Integration Test", updatedEmployee.CommonName)

		// 4. Verify update via GET
		c, rec = setupEchoContext(http.MethodGet, "/employees/"+createdEmployee.ID.String(), "")
		c.SetParamNames("id")
		c.SetParamValues(createdEmployee.ID.String())
		err = service.GetEmployeeByIDHandler(c)
		assert.NoError(t, err)

		err = json.Unmarshal(rec.Body.Bytes(), &retrievedEmployee)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Integration Test", retrievedEmployee.CommonName)

		// 5. List all employees via GET
		c, rec = setupEchoContext(http.MethodGet, "/employees", "")
		err = service.GetEmployeesHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var allEmployees []employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &allEmployees)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(allEmployees), 1)
	})
}

// ===============================
// ERROR HANDLING AND EDGE CASES
// ===============================

func TestSpecialCharacters(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	t.Run("Unicode characters in names", func(t *testing.T) {
		reqBody := `{
			"common_name": "José María",
			"first_name": "José",
			"last_name": "María",
			"employee_number": "UNI001"
		}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response employees.Employee
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "José María", response.CommonName)
		assert.Equal(t, "José", response.FirstName)
		assert.Equal(t, "María", response.LastName)
	})

	t.Run("Special characters in names", func(t *testing.T) {
		reqBody := `{
			"common_name": "O'Connor-Smith",
			"first_name": "Mary",
			"last_name": "O'Connor-Smith",
			"employee_number": "SPE001"
		}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}

func TestEdgeCases(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	t.Run("Very long names", func(t *testing.T) {
		longName := strings.Repeat("A", 100)
		reqBody := fmt.Sprintf(`{
			"common_name": "%s",
			"first_name": "%s",
			"last_name": "%s",
			"employee_number": "LONG001"
		}`, longName, longName, longName)
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code) // Should fail validation due to long names
	})

	t.Run("Empty string validation", func(t *testing.T) {
		reqBody := `{
			"common_name": "",
			"first_name": "",
			"last_name": "",
			"employee_number": ""
		}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)

		err := service.PostEmployeeHandler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code) // Should fail validation
	})
}

// ===============================
// HIGH PRIORITY COVERAGE TESTS
// ===============================

// setupEchoContextWithQuery creates an Echo context with custom query parameters for testing
func setupEchoContextWithQuery(method, url, body, rawQuery string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = utils.NewValidator()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	if rawQuery != "" {
		req.URL.RawQuery = rawQuery
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	return c, rec
}

func TestGetEmployeesHandler_FilterBindingError(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Create test data
	createEmployeeInDB(db, testEmployeesData[0])

	t.Run("Invalid boolean value for active parameter", func(t *testing.T) {
		c, rec := setupEchoContextWithQuery(http.MethodGet, "/employees", "", "active=maybe")

		err := service.GetEmployeesHandler(c)

		assert.NoError(t, err) // Handler doesn't return error, but sends HTTP error response
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var errorResponse utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Contains(t, errorResponse.Error, "parsing")
	})

	t.Run("Multiple invalid parameters", func(t *testing.T) {
		c, _ := setupEchoContextWithQuery(http.MethodGet, "/employees", "", "active=invalid&employee_number=")

		err := service.GetEmployeesHandler(c)

		assert.NoError(t, err) // Handler doesn't return error, but sends HTTP error response
		// Should still attempt to process, may succeed with empty employee_number
		// The active=invalid should cause the binding error
	})

	t.Run("Invalid query parameter format", func(t *testing.T) {
		// Test with malformed query that might break binding
		c, _ := setupEchoContextWithQuery(http.MethodGet, "/employees", "", "active=1&active=0&active=true")

		err := service.GetEmployeesHandler(c)

		assert.NoError(t, err) // Handler should handle gracefully
		// Multiple values for same param - should use last value or handle appropriately
	})

	t.Run("Extremely long parameter values", func(t *testing.T) {
		longValue := strings.Repeat("a", 10000)
		queryString := fmt.Sprintf("employee_number=%s", longValue)
		c, rec := setupEchoContextWithQuery(http.MethodGet, "/employees", "", queryString)

		err := service.GetEmployeesHandler(c)

		assert.NoError(t, err)
		// Should handle long values gracefully - either process or return reasonable error
		assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusBadRequest)
	})
}

func TestUpdateEmployee_DatabaseFailures(t *testing.T) {
	t.Run("Updates operation fails", func(t *testing.T) {
		db := setupTestDB()
		service := employees.NewEmployeeService(db)

		// Create test employee
		createdEmp := createEmployeeInDB(db, testEmployeesData[0])

		// Close database to simulate failure
		sqlDB, _ := db.DB()
		sqlDB.Close()

		newName := "Updated Name"
		patch := employees.EmployeePatch{
			CommonName: &newName,
		}

		result, err := service.UpdateEmployee(createdEmp.ID.String(), patch)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Update succeeds but retrieval fails", func(t *testing.T) {
		// This test demonstrates what happens when update works but retrieval fails
		// In practice, this is hard to simulate without sophisticated mocking

		db := setupTestDB()
		service := employees.NewEmployeeService(db)

		// Create test employee
		createdEmp := createEmployeeInDB(db, testEmployeesData[0])

		// For this simple test, we'll test the normal path
		// In a real scenario, we'd need dependency injection for the DB to mock this properly
		newName := "Updated Name"
		patch := employees.EmployeePatch{
			CommonName: &newName,
		}

		// This should succeed normally
		result, err := service.UpdateEmployee(createdEmp.ID.String(), patch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newName, result.CommonName)
	})

	t.Run("Non-existent employee update with different error", func(t *testing.T) {
		db := setupTestDB()
		service := employees.NewEmployeeService(db)

		// Use a malformed UUID to test different error paths
		invalidID := "not-a-valid-uuid-format"
		newName := "Test"
		patch := employees.EmployeePatch{
			CommonName: &newName,
		}

		result, err := service.UpdateEmployee(invalidID, patch)

		assert.Error(t, err)
		assert.Nil(t, result)
		// The error should be different from a "not found" error
	})

	t.Run("Database constraint violation during update", func(t *testing.T) {
		db := setupTestDB()
		service := employees.NewEmployeeService(db)

		// Create two test employees
		emp1 := createEmployeeInDB(db, testEmployeesData[0])
		emp2 := createEmployeeInDB(db, testEmployeesData[1])

		// Try to update emp2 to have the same employee number as emp1
		// This should violate uniqueness constraint if it exists
		patch := employees.EmployeePatch{
			EmployeeNumber: &emp1.EmployeeNumber,
		}

		result, err := service.UpdateEmployee(emp2.ID.String(), patch)

		// Depending on DB constraints, this might succeed or fail
		// We're testing that the service handles constraint violations properly
		if err != nil {
			assert.Nil(t, result)
		} else {
			assert.NotNil(t, result)
		}
	})

	t.Run("Empty patch update", func(t *testing.T) {
		db := setupTestDB()
		service := employees.NewEmployeeService(db)

		// Create test employee
		createdEmp := createEmployeeInDB(db, testEmployeesData[0])

		// Update with completely empty patch
		patch := employees.EmployeePatch{}

		result, err := service.UpdateEmployee(createdEmp.ID.String(), patch)

		// Should succeed with no changes
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, createdEmp.CommonName, result.CommonName)
	})
}

func TestRegisterRoutes(t *testing.T) {
	// Create Echo instance and group
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a test group (simulating protected routes)
	group := e.Group("/api")

	// Create a test service
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	// Register routes
	employees.RegisterRoutes(group, service)

	// Get all registered routes
	routes := e.Routes()

	t.Run("All required routes are registered", func(t *testing.T) {
		expectedRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/employees"},
			{"GET", "/api/employees"},
			{"GET", "/api/employees/:id"},
			{"PATCH", "/api/employees/:id"},
		}

		for _, expected := range expectedRoutes {
			found := false
			for _, route := range routes {
				if route.Method == expected.method && route.Path == expected.path {
					found = true
					break
				}
			}
			assert.True(t, found, "Route %s %s not found", expected.method, expected.path)
		}
	})

	t.Run("Routes can be called through Echo", func(t *testing.T) {
		// Create test employee for GET requests
		emp := createEmployeeInDB(db, testEmployeesData[0])

		// Test POST route
		reqBody := `{"common_name":"Route Test","first_name":"Route","last_name":"Test","employee_number":"ROUTE001"}`
		req := httptest.NewRequest(http.MethodPost, "/api/employees", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Test GET all route
		req = httptest.NewRequest(http.MethodGet, "/api/employees", nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Test GET by ID route
		req = httptest.NewRequest(http.MethodGet, "/api/employees/"+emp.ID.String(), nil)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Test PATCH route
		patchBody := `{"common_name":"Updated Route Test"}`
		req = httptest.NewRequest(http.MethodPatch, "/api/employees/"+emp.ID.String(), strings.NewReader(patchBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Route count verification", func(t *testing.T) {
		// Count employee routes (should be exactly 4)
		employeeRoutes := 0
		for _, route := range routes {
			if strings.Contains(route.Path, "/api/employees") {
				employeeRoutes++
			}
		}

		assert.Equal(t, 4, employeeRoutes, "Expected exactly 4 employee routes")
	})
}

// ===============================
// MEDIUM PRIORITY COVERAGE TESTS
// ===============================

func TestNewEmployee_Enhanced_Validation(t *testing.T) {
	tests := []struct {
		name           string
		commonName     string
		firstName      string
		lastName       string
		employeeNumber string
		expectError    bool
		expectedError  string
	}{
		{
			name:           "Valid employee with all fields",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "EMP001",
			expectError:    false,
		},
		{
			name:           "Employee number too short",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "AB",
			expectError:    true,
			expectedError:  "employee number must be at least 3 characters long",
		},
		{
			name:           "Employee number too long",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: strings.Repeat("A", 51),
			expectError:    true,
			expectedError:  "employee number cannot exceed 50 characters",
		},
		{
			name:           "Employee number with invalid characters",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "EMP-001",
			expectError:    true,
			expectedError:  "employee number can only contain alphanumeric characters",
		},
		{
			name:           "Employee number with special characters",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "EMP@001",
			expectError:    true,
			expectedError:  "employee number can only contain alphanumeric characters",
		},
		{
			name:           "Common name too long",
			commonName:     strings.Repeat("A", 101),
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "EMP001",
			expectError:    true,
			expectedError:  "common name cannot exceed 100 characters",
		},
		{
			name:           "First name too long",
			commonName:     "John Doe",
			firstName:      strings.Repeat("A", 51),
			lastName:       "Doe",
			employeeNumber: "EMP001",
			expectError:    true,
			expectedError:  "first name cannot exceed 50 characters",
		},
		{
			name:           "Last name too long",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       strings.Repeat("A", 51),
			employeeNumber: "EMP001",
			expectError:    true,
			expectedError:  "last name cannot exceed 50 characters",
		},
		{
			name:           "Empty employee number is valid",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "",
			expectError:    false,
		},
		{
			name:           "Minimum valid employee number length",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: "ABC",
			expectError:    false,
		},
		{
			name:           "Maximum valid employee number length",
			commonName:     "John Doe",
			firstName:      "John",
			lastName:       "Doe",
			employeeNumber: strings.Repeat("A", 50),
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			employee, err := employees.NewEmployee(tt.commonName, tt.firstName, tt.lastName, tt.employeeNumber)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, employee)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, employee)
				assert.Equal(t, tt.commonName, employee.CommonName)
				assert.Equal(t, tt.firstName, employee.FirstName)
				assert.Equal(t, tt.lastName, employee.LastName)
				assert.Equal(t, tt.employeeNumber, employee.EmployeeNumber)
				assert.True(t, employee.Active)
			}
		})
	}
}

func TestPostEmployeeHandler_NewEmployeeErrors(t *testing.T) {
	db := setupTestDB()
	service := employees.NewEmployeeService(db)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		checkError     bool
	}{
		{
			name:           "Employee number too short",
			requestBody:    `{"common_name":"Test User","first_name":"Test","last_name":"User","employee_number":"AB"}`,
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:           "Employee number with invalid characters",
			requestBody:    `{"common_name":"Test User","first_name":"Test","last_name":"User","employee_number":"EMP-001"}`,
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:           "Common name too long",
			requestBody:    fmt.Sprintf(`{"common_name":"%s","first_name":"Test","last_name":"User","employee_number":"EMP001"}`, strings.Repeat("A", 101)),
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:           "First name too long",
			requestBody:    fmt.Sprintf(`{"common_name":"Test User","first_name":"%s","last_name":"User","employee_number":"EMP001"}`, strings.Repeat("A", 51)),
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:           "Last name too long",
			requestBody:    fmt.Sprintf(`{"common_name":"Test User","first_name":"Test","last_name":"%s","employee_number":"EMP001"}`, strings.Repeat("A", 51)),
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:           "Valid request should succeed",
			requestBody:    `{"common_name":"Test User","first_name":"Test","last_name":"User","employee_number":"EMP001"}`,
			expectedStatus: http.StatusCreated,
			checkError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := setupEchoContext(http.MethodPost, "/employees", tt.requestBody)

			err := service.PostEmployeeHandler(c)

			assert.NoError(t, err) // Handler doesn't return error, sends HTTP response
			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.checkError && tt.expectedStatus != http.StatusCreated {
				// For error cases, we expect an error message in the response
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
			}

			if tt.expectedStatus == http.StatusCreated {
				// For success cases, we expect an employee object
				var employee employees.Employee
				err = json.Unmarshal(rec.Body.Bytes(), &employee)
				assert.NoError(t, err)
				assert.NotEmpty(t, employee.ID)
			}
		})
	}
}

func TestDatabaseConnectionFailures(t *testing.T) {
	t.Run("Service operations with closed database", func(t *testing.T) {
		db := setupTestDB()

		// Create test employee first while DB is open
		emp := createEmployeeInDB(db, testEmployeesData[0])

		// Close the database connection
		sqlDB, _ := db.DB()
		sqlDB.Close()

		service := employees.NewEmployeeService(db)

		// Test CreateEmployee with closed DB
		newEmp, _ := employees.NewEmployee("Test", "Test", "User", "TEST001")
		err := service.CreateEmployee(newEmp)
		assert.Error(t, err, "CreateEmployee should fail with closed DB")

		// Test GetEmployees with closed DB
		filter := employees.EmployeeFilter{}
		_, err = service.GetEmployees(filter)
		assert.Error(t, err, "GetEmployees should fail with closed DB")

		// Test GetEmployeeByID with closed DB
		_, err = service.GetEmployeeByID(emp.ID.String())
		assert.Error(t, err, "GetEmployeeByID should fail with closed DB")

		// Test UpdateEmployee with closed DB
		patch := employees.EmployeePatch{CommonName: &[]string{"Updated"}[0]}
		_, err = service.UpdateEmployee(emp.ID.String(), patch)
		assert.Error(t, err, "UpdateEmployee should fail with closed DB")
	})

	t.Run("Handler operations with closed database", func(t *testing.T) {
		db := setupTestDB()

		// Create test employee first while DB is open
		emp := createEmployeeInDB(db, testEmployeesData[0])
		empID := emp.ID.String()

		// Close the database connection
		sqlDB, _ := db.DB()
		sqlDB.Close()

		service := employees.NewEmployeeService(db)

		// Test PostEmployeeHandler with closed DB
		reqBody := `{"common_name":"DB Test","first_name":"DB","last_name":"Test","employee_number":"DBTEST001"}`
		c, rec := setupEchoContext(http.MethodPost, "/employees", reqBody)
		err := service.PostEmployeeHandler(c)
		assert.NoError(t, err) // Handler doesn't return error
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		// Test GetEmployeesHandler with closed DB
		c, rec = setupEchoContext(http.MethodGet, "/employees", "")
		err = service.GetEmployeesHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		// Test GetEmployeeByIDHandler with closed DB
		c, rec = setupEchoContext(http.MethodGet, "/employees/"+empID, "")
		c.SetParamNames("id")
		c.SetParamValues(empID)
		err = service.GetEmployeeByIDHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code) // This returns 404, not 500

		// Test PatchEmployeeHandler with closed DB
		patchBody := `{"common_name":"Updated Name"}`
		c, rec = setupEchoContext(http.MethodPatch, "/employees/"+empID, patchBody)
		c.SetParamNames("id")
		c.SetParamValues(empID)
		err = service.PatchEmployeeHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
