.PHONY: swagger build run test clean dev lint perf-test bench stress-test profile

swagger:
	swag init --parseDependency --parseInternal -g cmd/qc_api/main.go

build: swagger
	go build -o ./tmp/main cmd/qc_api/main.go

run: build
	./tmp/main

test:
	go test ./...

dev:
	air

clean:
	rm -rf tmp/ docs/

lint:
	golangci-lint run

# Performance testing targets
perf-test:
	@echo "Running performance tests..."
	go test -v -run="Performance" ./internal/employees/

bench:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./internal/employees/

stress-test:
	@echo "Running stress tests..."
	go test -v -run="Stress" ./internal/employees/

profile:
	@echo "Starting profiling server..."
	@echo "Visit http://localhost:6060/debug/pprof/"
	@echo "Use 'go tool pprof http://localhost:6060/debug/pprof/profile' for CPU profiling"
	@echo "Use 'go tool pprof http://localhost:6060/debug/pprof/heap' for memory profiling"

# Combined performance testing
perf-all: perf-test bench stress-test
	@echo "All performance tests completed!"

# Quick performance check (faster subset)
perf-quick:
	@echo "Running quick performance tests..."
	go test -v -run="PerformanceConcurrentEmployeeCreation" ./internal/employees/
	go test -bench="BenchmarkCreateEmployee$$|BenchmarkGetEmployeeByID$$" -benchmem ./internal/employees/
