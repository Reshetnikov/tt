package utils

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type FormErrors map[string][]string
func (fe *FormErrors) Add(field, message string) {
	(*fe)[field] = append((*fe)[field], message)
}
func (fe FormErrors) HasErrorsField(field string) bool {
	return len(fe[field]) > 0
}
func (fe FormErrors) HasErrors() bool {
	return len(fe) > 0
}

// 
func ValidateStruct(formStruct interface{}) FormErrors {
	errors := FormErrors{}
	if valErrors := validate.Struct(formStruct); valErrors != nil {
		for _, valError := range valErrors.(validator.ValidationErrors) {
			fieldName := valError.Field()
			fieldLabel := getFieldLabel(formStruct, fieldName)
			tag := valError.Tag()
			errorMessage := parseValidationError(tag, valError, fieldLabel)
			errors.Add(fieldName, errorMessage)
		}
	}
	return errors
}

// Receives a human message
func parseValidationError(tag string, err validator.FieldError, fieldLabel string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", fieldLabel)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fieldLabel)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fieldLabel, err.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", fieldLabel, err.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", fieldLabel, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", fieldLabel)
	}
}

// Gets the label for the field or use the default field name
// Example: PasswordConfirmation string `label:"Confirm Password"`
func getFieldLabel(formStruct interface{}, fieldName string) string {
	val := reflect.ValueOf(formStruct)
	field, found := val.Type().FieldByName(fieldName)
	if !found {
		return fieldName
	}
	label := field.Tag.Get("label")
	if label == "" {
		return fieldName
	}
	return label
}