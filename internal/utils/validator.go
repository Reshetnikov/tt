package utils

import (
	"fmt"
	"reflect"
	"strings"

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

// Use "formErrors.HasErrors()" instead of "formErrors == nil"
// because "formErrors := utils.FormErrors{}" is non-nil.
func (fe FormErrors) HasErrors() bool {
	return len(fe) > 0
}
func (fe FormErrors) Error() string {
	var sb strings.Builder
	for field, messages := range fe {
		sb.WriteString(fmt.Sprintf("Errors for field '%s': %s\n", field, strings.Join(messages, ", ")))
	}
	return sb.String()
}

// Example usage:
//
//	form := signupForm{}
//	validator := utils.NewValidator(&form)
//	errors := validator.Validate()
type Validator struct {
	formStruct    interface{}
	valFormStruct reflect.Value
	formErrors    FormErrors
}

func NewValidator(formStruct interface{}) *Validator {
	return &Validator{
		formStruct:    formStruct,
		valFormStruct: reflect.ValueOf(formStruct).Elem(),
		formErrors:    FormErrors{},
	}
}
func (v Validator) Validate() FormErrors {
	if valErrors := validate.Struct(v.formStruct); valErrors != nil {
		for _, valError := range valErrors.(validator.ValidationErrors) {
			fieldName := valError.Field()
			fieldLabel := v.getFieldLabel(fieldName)
			tag := valError.Tag()
			errorMessage := v.parseValidationError(tag, valError, fieldLabel)
			v.formErrors.Add(fieldName, errorMessage)
		}
	}
	return v.formErrors
}

// Receives a human message
func (v Validator) parseValidationError(tag string, err validator.FieldError, fieldLabel string) string {
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
		field2Label := v.getFieldLabel(err.Param())
		return fmt.Sprintf("%s must match %s", fieldLabel, field2Label)
	default:
		return fmt.Sprintf("%s is invalid", fieldLabel)
	}
}

// Gets the label for the field or use the default field name
// Example: PasswordConfirmation string `label:"Confirm Password"`
func (v Validator) getFieldLabel(fieldName string) string {
	field, found := v.valFormStruct.Type().FieldByName(fieldName)
	if !found {
		return fieldName
	}
	label := field.Tag.Get("label")
	if label == "" {
		return fieldName
	}
	return label
}
