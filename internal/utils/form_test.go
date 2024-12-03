//go:build unit

package utils

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name     string `form:"name"`
	Age      int    `form:"age"`
	IsActive bool   `form:"is_active"`
	Hidden   string `form:"-"`
}
type StructWithNoFormTags struct {
	Field1 string
	Field2 int
}

func createMockRequest(values url.Values) *http.Request {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.PostForm = values
	return req
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -run TestParseFormToStruct.*
func TestParseFormToStruct(t *testing.T) {
	testCases := []struct {
		name           string
		formValues     url.Values
		expectedStruct TestStruct
	}{
		{
			name: "All fields filled",
			formValues: url.Values{
				"name":      {"John Doe"},
				"age":       {"30"},
				"is_active": {"on"},
			},
			expectedStruct: TestStruct{
				Name:     "John Doe",
				Age:      30,
				IsActive: true,
			},
		},
		{
			name: "Partial fields",
			formValues: url.Values{
				"name": {"Jane Doe"},
			},
			expectedStruct: TestStruct{
				Name:     "Jane Doe",
				Age:      0,
				IsActive: false,
			},
		},
		{
			name:           "Empty form",
			formValues:     url.Values{},
			expectedStruct: TestStruct{},
		},
		{
			name: "Boolean false cases",
			formValues: url.Values{
				"is_active": {"false"},
			},
			expectedStruct: TestStruct{
				IsActive: false,
			},
		},
		{
			name: "Invalid integer",
			formValues: url.Values{
				"age": {"not a number"},
			},
			expectedStruct: TestStruct{}, // Age should remain 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := createMockRequest(tc.formValues)
			actualStruct := TestStruct{}

			err := ParseFormToStruct(req, &actualStruct)

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStruct, actualStruct)
		})
	}
}

func TestParseFormToStruct_StructWithNoFormTags(t *testing.T) {
	req := createMockRequest(url.Values{
		"Field1": {"test"},
		"Field2": {"42"},
	})
	actualStruct := StructWithNoFormTags{}

	err := ParseFormToStruct(req, &actualStruct)

	assert.NoError(t, err)
	assert.Equal(t, StructWithNoFormTags{}, actualStruct)
}

func TestParseFormToStruct_InvalidInput(t *testing.T) {
	req, _ := http.NewRequest("POST", "/test", nil)
	req.Form = nil // Simulate error in ParseForm()

	actualStruct := TestStruct{}
	err := ParseFormToStruct(req, &actualStruct)

	assert.Error(t, err)
}
