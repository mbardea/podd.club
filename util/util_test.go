package util

import (
	"net/http"
	"testing"
)

func TestRangeHeader(t *testing.T) {
	// headers := make(map[string]string)
	headers := &http.Header{}
	headers.Add("Range", "[bytes=100-200]")
	start, end := ParseRangeHeader(headers, 1, 2)
	expectedStart := int64(100)
	if start != expectedStart {
		t.Errorf("Failed to parse start. Expected %d, Got %d", expectedStart, start)
	}
	expectedEnd := int64(200)
	if end != expectedEnd {
		t.Errorf("Failed to parse end. Expected %d, Got %d", expectedEnd, end)
	}
}
