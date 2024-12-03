//go:build unit

package utils

import (
	"testing"
)

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestStr.*
func TestStrColor(t *testing.T) {
	tests := []struct {
		input    string
		color    string
		expected string
	}{
		{"Hello", Colors.Blue, "\033[1;34mHello\033[0m"},
		{"Warning", Colors.Yellow, "\033[1;33mWarning\033[0m"},
		{"Success", Colors.Green, "\033[1;32mSuccess\033[0m"},
		{"Error", Colors.Red, "\033[31mError\033[0m"},
	}

	for _, tt := range tests {
		result := StrColor(tt.input, tt.color)
		if result != tt.expected {
			t.Errorf("StrColor(%q, %q) = %q; want %q", tt.input, tt.color, result, tt.expected)
		}
	}
}
