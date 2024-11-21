package utils

import (
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
)

// Fills the structure with data from the form
func ParseFormToStruct(r *http.Request, formStruct interface{}) error {
	if err := r.ParseForm(); err != nil {
		slog.Error("ParseFormToStruct ParseForm", "err", err)
		return err
	}

	valFormStruct := reflect.ValueOf(formStruct).Elem()

	for i := 0; i < valFormStruct.NumField(); i++ {
		fieldStruct := valFormStruct.Type().Field(i)
		fieldForm := fieldStruct.Tag.Get("form")
		if fieldForm == "" {
			continue
		}

		if values, ok := r.Form[fieldForm]; ok {
			if valFormStruct.Field(i).CanSet() {
				switch valFormStruct.Field(i).Kind() {
				case reflect.String:
					valFormStruct.Field(i).SetString(values[0])
				case reflect.Bool:
					valFormStruct.Field(i).SetBool(values[0] == "on" || values[0] == "true")
				case reflect.Int, reflect.Int64:
					if intValue, err := strconv.Atoi(values[0]); err == nil {
						valFormStruct.Field(i).SetInt(int64(intValue))
					}
				default:
					//
				}
			}
		} else {
			// If the field is missing from the form, we handle the specific case for bool
			if valFormStruct.Field(i).Kind() == reflect.Bool && valFormStruct.Field(i).CanSet() {
				valFormStruct.Field(i).SetBool(false)
			}
		}
	}

	return nil
}
