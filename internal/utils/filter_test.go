package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestRecord struct {
	ID        uint      `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
	Name      string
}

type TestFilter struct {
	DateFrom *SimpleDate `json:"date_from,omitempty" query:"date_from" filter:"created_at"`
	DateTo   *SimpleDate `json:"date_to,omitempty" query:"date_to" filter:"created_at"`
	Name     *string     `json:"name,omitempty" query:"name"`
}

func TestApplyFilter_SimpleDate(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create table
	err = db.AutoMigrate(&TestRecord{})
	assert.NoError(t, err)

	// Insert test data
	records := []TestRecord{
		{ID: 1, CreatedAt: time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC), Name: "record1"},
		{ID: 2, CreatedAt: time.Date(2025, 9, 2, 12, 0, 0, 0, time.UTC), Name: "record2"},
		{ID: 3, CreatedAt: time.Date(2025, 9, 3, 12, 0, 0, 0, time.UTC), Name: "record3"},
	}

	for _, record := range records {
		err = db.Create(&record).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name          string
		filter        TestFilter
		expectedCount int
		expectedIDs   []uint
	}{
		{
			name: "Filter from date",
			filter: TestFilter{
				DateFrom: &SimpleDate{Time: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)},
			},
			expectedCount: 2,
			expectedIDs:   []uint{2, 3},
		},
		{
			name: "Filter to date",
			filter: TestFilter{
				DateTo: &SimpleDate{Time: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)},
			},
			expectedCount: 2,
			expectedIDs:   []uint{1, 2},
		},
		{
			name: "Filter date range",
			filter: TestFilter{
				DateFrom: &SimpleDate{Time: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)},
				DateTo:   &SimpleDate{Time: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)},
			},
			expectedCount: 1,
			expectedIDs:   []uint{2},
		},
		{
			name:          "No filter",
			filter:        TestFilter{},
			expectedCount: 3,
			expectedIDs:   []uint{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []TestRecord
			query := ApplyFilter(db.Model(&TestRecord{}), tt.filter)
			err := query.Find(&results).Error
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(results))

			var actualIDs []uint
			for _, result := range results {
				actualIDs = append(actualIDs, result.ID)
			}
			assert.ElementsMatch(t, tt.expectedIDs, actualIDs)
		})
	}
}
