//go:build unit

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestForm struct {
	Name            string `validate:"required,min=3,max=20" label:"Name"`
	Email           string `validate:"required,email" label:"Email"`
	Password        string `validate:"required,min=8" label:"Password"`
	PasswordConfirm string `validate:"required,eqfield=Password" label:"Confirm Password"`
	// depends on non-existent field password2
	PasswordConfirm2 string `validate:"omitempty,eqfield=Password2" label:"Confirm Password2"`
	Color            string `form:"color" validate:"omitempty,hexcolor"`
}

// docker exec -it tt-app-1 go test -v ./internal/utils --tags=unit -cover -run TestValidator.*
func TestValidator_Validate(t *testing.T) {
	testCases := []struct {
		name            string
		form            TestForm
		expected        FormErrors
		hasErrorsFields []string
		hasErrors       bool
	}{
		{
			name: "valid form",
			form: TestForm{
				Name:            "John",
				Email:           "john.doe@example.com",
				Password:        "password123",
				PasswordConfirm: "password123",
				Color:           "#FFFFFF",
			},
			expected:        FormErrors{},
			hasErrorsFields: []string{},
			hasErrors:       false,
		},
		{
			name: "missing required fields",
			form: TestForm{
				Name:            "",
				Email:           "",
				Password:        "",
				PasswordConfirm: "",
			},
			expected: FormErrors{
				"Name":            {"Name is required"},
				"Email":           {"Email is required"},
				"Password":        {"Password is required"},
				"PasswordConfirm": {"Confirm Password is required"},
			},
			hasErrorsFields: []string{"Name", "Email", "Password", "PasswordConfirm"},
			hasErrors:       true,
		},
		{
			name: "invalid fields",
			form: TestForm{
				Name:             "Jooooooooooooooooooooo",
				Email:            "not-an-email",
				Password:         "pass123",
				PasswordConfirm:  "pass456",
				PasswordConfirm2: "pass111",
				Color:            "red",
			},
			expected: FormErrors{
				"Name":             {"Name must not exceed 20 characters"},
				"Email":            {"Email must be a valid email address"},
				"Password":         {"Password must be at least 8 characters"},
				"PasswordConfirm":  {"Confirm Password must match Password"},
				"PasswordConfirm2": {"Confirm Password2 must match Password2"},
				"Color":            {"Color is invalid"},
			},
			hasErrorsFields: []string{"Name", "Email", "Password", "PasswordConfirm"},
			hasErrors:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := NewValidator(&tc.form)
			errors := validator.Validate()

			if tc.hasErrors {
				assert.True(t, errors.HasErrors())
				assert.True(t, len(errors.Error()) > 0)
			} else {
				assert.False(t, errors.HasErrors())
				assert.True(t, len(errors.Error()) == 0)
			}

			for _, field := range tc.hasErrorsFields {
				assert.True(t, errors.HasErrorsField(field))
			}
			assert.Equal(t, tc.expected, errors)
		})
	}
}
