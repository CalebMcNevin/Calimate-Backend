package employees_test

import (
	"fmt"
	"testing"

	"qc_api/internal/employees"
	"qc_api/internal/utils"

	"github.com/stretchr/testify/require"
)

var benchmarkDB = utils.SetupPerformanceTestDB()

func init() {
	err := benchmarkDB.AutoMigrate(employees.Models()...)
	if err != nil {
		panic("failed to migrate benchmark database: " + err.Error())
	}
}

func BenchmarkCreateEmployee(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("BenchCreate_%d", i),
			fmt.Sprintf("Bench_%d", i),
			fmt.Sprintf("Create_%d", i),
			fmt.Sprintf("BC%07d", i),
		)
		if err != nil {
			b.Fatal(err)
		}

		err = service.CreateEmployee(emp)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEmployeeByID(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	// Setup: Create test employees
	const numTestEmployees = 1000
	employeeIDs := make([]string, numTestEmployees)

	for i := 0; i < numTestEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("BenchRead_%d", i),
			fmt.Sprintf("Bench_%d", i),
			fmt.Sprintf("Read_%d", i),
			fmt.Sprintf("BR%07d", i),
		)
		require.NoError(b, err)
		require.NoError(b, service.CreateEmployee(emp))
		employeeIDs[i] = emp.ID.String()
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		employeeID := employeeIDs[i%len(employeeIDs)]
		_, err := service.GetEmployeeByID(employeeID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEmployees(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)
	filter := employees.EmployeeFilter{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetEmployees(filter)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEmployeesWithFilter(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	// Setup: Ensure we have both active and inactive employees
	activeEmployee, err := employees.NewEmployee("Active Test", "Active", "Test", "ACTIVE01")
	require.NoError(b, err)
	require.NoError(b, service.CreateEmployee(activeEmployee))

	active := true
	filter := employees.EmployeeFilter{Active: &active}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetEmployees(filter)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateEmployee(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	// Setup: Create test employees for updating
	const numTestEmployees = 100
	employeeIDs := make([]string, numTestEmployees)

	for i := 0; i < numTestEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("BenchUpdate_%d", i),
			fmt.Sprintf("Bench_%d", i),
			fmt.Sprintf("Update_%d", i),
			fmt.Sprintf("BU%07d", i),
		)
		require.NoError(b, err)
		require.NoError(b, service.CreateEmployee(emp))
		employeeIDs[i] = emp.ID.String()
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		employeeID := employeeIDs[i%len(employeeIDs)]
		newName := fmt.Sprintf("Updated_%d", i)
		patch := employees.EmployeePatch{
			CommonName: &newName,
		}

		_, err := service.UpdateEmployee(employeeID, patch)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewEmployee(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := employees.NewEmployee(
			fmt.Sprintf("BenchNew_%d", i),
			fmt.Sprintf("Bench_%d", i),
			fmt.Sprintf("New_%d", i),
			fmt.Sprintf("BN%07d", i),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewEmployeeValidation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Test various validation scenarios
		switch i % 4 {
		case 0:
			// Valid employee
			_, err := employees.NewEmployee("Valid Name", "Valid", "Name", "VALID123")
			if err != nil {
				b.Fatal("Valid employee should not fail")
			}
		case 1:
			// Invalid - employee number too short
			_, err := employees.NewEmployee("Valid Name", "Valid", "Name", "AB")
			if err == nil {
				b.Fatal("Should have failed validation")
			}
		case 2:
			// Invalid - employee number with special characters
			_, err := employees.NewEmployee("Valid Name", "Valid", "Name", "EMP-123")
			if err == nil {
				b.Fatal("Should have failed validation")
			}
		case 3:
			// Invalid - name too long
			longName := string(make([]byte, 101))
			for j := range longName {
				longName = longName[:j] + "A" + longName[j+1:]
			}
			_, err := employees.NewEmployee(longName, "Valid", "Name", "VALID456")
			if err == nil {
				b.Fatal("Should have failed validation")
			}
		}
	}
}

// Benchmark parallel operations
func BenchmarkCreateEmployeeParallel(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			emp, err := employees.NewEmployee(
				fmt.Sprintf("ParallelCreate_%d", i),
				fmt.Sprintf("Parallel_%d", i),
				fmt.Sprintf("Create_%d", i),
				fmt.Sprintf("PC%07d", i),
			)
			if err != nil {
				b.Fatal(err)
			}

			err = service.CreateEmployee(emp)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkGetEmployeeByIDParallel(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	// Setup: Create test employees
	const numTestEmployees = 100
	employeeIDs := make([]string, numTestEmployees)

	for i := 0; i < numTestEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("ParallelRead_%d", i),
			fmt.Sprintf("Parallel_%d", i),
			fmt.Sprintf("Read_%d", i),
			fmt.Sprintf("PR%07d", i),
		)
		require.NoError(b, err)
		require.NoError(b, service.CreateEmployee(emp))
		employeeIDs[i] = emp.ID.String()
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			employeeID := employeeIDs[i%len(employeeIDs)]
			_, err := service.GetEmployeeByID(employeeID)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkUpdateEmployeeParallel(b *testing.B) {
	service := employees.NewEmployeeService(benchmarkDB)

	// Setup: Create test employees for updating
	const numTestEmployees = 50
	employeeIDs := make([]string, numTestEmployees)

	for i := 0; i < numTestEmployees; i++ {
		emp, err := employees.NewEmployee(
			fmt.Sprintf("ParallelUpdate_%d", i),
			fmt.Sprintf("Parallel_%d", i),
			fmt.Sprintf("Update_%d", i),
			fmt.Sprintf("PU%07d", i),
		)
		require.NoError(b, err)
		require.NoError(b, service.CreateEmployee(emp))
		employeeIDs[i] = emp.ID.String()
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			employeeID := employeeIDs[i%len(employeeIDs)]
			newName := fmt.Sprintf("ParallelUpdated_%d", i)
			patch := employees.EmployeePatch{
				CommonName: &newName,
			}

			_, err := service.UpdateEmployee(employeeID, patch)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// Benchmark memory allocation patterns
func BenchmarkEmployeeAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		emp := &employees.Employee{
			CommonName:     fmt.Sprintf("Alloc_%d", i),
			FirstName:      fmt.Sprintf("First_%d", i),
			LastName:       fmt.Sprintf("Last_%d", i),
			EmployeeNumber: fmt.Sprintf("ALLOC%04d", i),
			Active:         true,
		}
		_ = emp // Use the variable to prevent optimization
	}
}

func BenchmarkEmployeeDTOAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dto := employees.EmployeeDTO{
			CommonName:     fmt.Sprintf("DTO_%d", i),
			FirstName:      fmt.Sprintf("First_%d", i),
			LastName:       fmt.Sprintf("Last_%d", i),
			EmployeeNumber: fmt.Sprintf("DTO%04d", i),
		}
		_ = dto // Use the variable to prevent optimization
	}
}

func BenchmarkEmployeePatchAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		commonName := fmt.Sprintf("Patch_%d", i)
		firstName := fmt.Sprintf("First_%d", i)
		active := true

		patch := employees.EmployeePatch{
			CommonName: &commonName,
			FirstName:  &firstName,
			Active:     &active,
		}
		_ = patch // Use the variable to prevent optimization
	}
}
