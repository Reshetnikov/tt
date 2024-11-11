package utils

import (
	"log/slog"
	"net/http"
	"reflect"
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

		if values, ok := r.Form[fieldForm]; ok {
			if valFormStruct.Field(i).CanSet() {
				valFormStruct.Field(i).SetString(values[0])
			}
		}
	}

	return nil
}
