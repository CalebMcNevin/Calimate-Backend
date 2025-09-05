package utils

import (
	"strings"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SimpleDate struct {
	time.Time
}

func (sd *SimpleDate) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" || str == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02", str)
	if err != nil {
		return err
	}

	sd.Time = t
	return nil
}

func (sd SimpleDate) MarshalJSON() ([]byte, error) {
	if sd.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + sd.Time.Format("2006-01-02") + `"`), nil
}

// UnmarshalParam implements Echo's param unmarshaling interface for query parameters
func (sd *SimpleDate) UnmarshalParam(param string) error {
	if param == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02", param)
	if err != nil {
		return err
	}

	sd.Time = t
	return nil
}
