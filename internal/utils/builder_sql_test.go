//go:build unit

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -run ^TestBuild.*$
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -run TestBuilderUpdate_BuildFromArr
func TestBuilderUpdate_BuildFromArr(t *testing.T) {
	testCases := []struct {
		name           string
		input          Arr
		expectedQuery  string
		expectedParams []interface{}
	}{
		{
			name: "Single field",
			input: Arr{
				{"name", "John Doe"},
			},
			expectedQuery:  "name = $1",
			expectedParams: []interface{}{"John Doe"},
		},
		{
			name: "Multiple fields",
			input: Arr{
				{"name", "John Doe"},
				{"email", "john@example.com"},
				{"age", 30},
			},
			expectedQuery:  "name = $1, email = $2, age = $3",
			expectedParams: []interface{}{"John Doe", "john@example.com", 30},
		},
		{
			name:           "Empty array",
			input:          Arr{},
			expectedQuery:  "",
			expectedParams: []interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := NewBuilderFieldsValues()
			query := builder.BuildFromArr(tc.input)

			assert.Equal(t, tc.expectedQuery, query)
			assert.Equal(t, tc.expectedParams, builder.Params())
		})
	}
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -run TestBuildFieldsFromArr
func TestBuildFieldsFromArr(t *testing.T) {
	testCases := []struct {
		name                 string
		input                Arr
		expectedFields       string
		expectedPlaceholders string
		expectedParams       []interface{}
	}{
		{
			name: "Single field",
			input: Arr{
				{"name", "John Doe"},
			},
			expectedFields:       "name",
			expectedPlaceholders: "$1",
			expectedParams:       []interface{}{"John Doe"},
		},
		{
			name: "Multiple fields",
			input: Arr{
				{"name", "John Doe"},
				{"email", "john@example.com"},
				{"age", 30},
			},
			expectedFields:       "name, email, age",
			expectedPlaceholders: "$1, $2, $3",
			expectedParams:       []interface{}{"John Doe", "john@example.com", 30},
		},
		{
			name:                 "Empty array",
			input:                Arr{},
			expectedFields:       "",
			expectedPlaceholders: "",
			expectedParams:       []interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields, placeholders, params := BuildFieldsFromArr(tc.input)

			assert.Equal(t, tc.expectedFields, fields)
			assert.Equal(t, tc.expectedPlaceholders, placeholders)
			assert.Equal(t, tc.expectedParams, params)
		})
	}
}
