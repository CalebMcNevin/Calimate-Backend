package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleDate_UnmarshalParam(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		hasError bool
	}{
		{
			name:     "Valid date",
			input:    "2025-09-02",
			expected: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: time.Time{},
			hasError: false,
		},
		{
			name:     "Invalid format",
			input:    "2025/09/02",
			expected: time.Time{},
			hasError: true,
		},
		{
			name:     "Invalid date",
			input:    "2025-13-45",
			expected: time.Time{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := &SimpleDate{}
			err := sd.UnmarshalParam(tt.input)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if !tt.expected.IsZero() {
					assert.Equal(t, tt.expected.Year(), sd.Time.Year())
					assert.Equal(t, tt.expected.Month(), sd.Time.Month())
					assert.Equal(t, tt.expected.Day(), sd.Time.Day())
				}
			}
		})
	}
}

func TestSimpleDate_JSON(t *testing.T) {
	sd := SimpleDate{Time: time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)}

	// Test marshaling
	data, err := sd.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"2025-09-02"`, string(data))

	// Test unmarshaling
	var sd2 SimpleDate
	err = sd2.UnmarshalJSON([]byte(`"2025-09-02"`))
	assert.NoError(t, err)
	assert.Equal(t, sd.Time.Year(), sd2.Time.Year())
	assert.Equal(t, sd.Time.Month(), sd2.Time.Month())
	assert.Equal(t, sd.Time.Day(), sd2.Time.Day())
}
