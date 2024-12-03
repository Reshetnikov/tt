package utils

import (
	"fmt"
	"strings"
)

type Map map[string]interface{}
type Arr [][]interface{}

type BuilderUpdate struct {
	params     []interface{}
	paramIndex int
}

func NewBuilderFieldsValues() *BuilderUpdate {
	return &BuilderUpdate{
		paramIndex: 1,
	}
}

// Example:
//
//	set := builder.BuildFromArr(utils.Arr{
//		{"name",               user.Name},
//		{"password",           user.Password},
//		{"email",              user.Email},
//	})
//
// where := builder.BuildFromArr(utils.Arr{{"id", user.ID}})
// query := "UPDATE users SET " + set + " WHERE " + where
// rows.Scan(builder.Params()...)
// Rerult:
// UPDATE users SET name = $1, password = $2, email = $3 WHERE id = $4
// user.Name, user.Password, user.Email, user.ID
func (b *BuilderUpdate) BuildFromArr(fieldsValues Arr) string {
	fieldsPlaceholders := make([]string, 0, len(fieldsValues))
	for _, fieldValue := range fieldsValues {
		fieldsPlaceholders = append(fieldsPlaceholders, fmt.Sprintf("%s = $%d", fieldValue[0], b.paramIndex))
		b.params = append(b.params, fieldValue[1])
		b.paramIndex++
	}
	return strings.Join(fieldsPlaceholders, ", ")
}

func (b *BuilderUpdate) Params() []interface{} {
	if b.params == nil {
		return []interface{}{}
	}
	return b.params
}

// Example
//
//	fields, placeholders, params := utils.BuildFieldsFromArr(utils.Arr{
//		{"name",               user.Name},
//		{"password",           user.Password},
//		{"email",              user.Email},
//	})
//
// query := "INSERT INTO users (" + fields + ") VALUES (" + placeholders + ")"
// INSERT INTO users (name, password, email) VALUES ($1, $2, $3)
// db.Exec(query, params...) // user.Name, user.Password, user.Email
func BuildFieldsFromArr(fieldsValues Arr) (fields string, placeholders string, params []interface{}) {
	params = []interface{}{}
	fieldsArr := make([]string, 0, len(fieldsValues))
	placeholdersArr := make([]string, 0, len(fieldsValues))
	// params = []interface{}{}
	paramIndex := 1
	for _, fieldValue := range fieldsValues {
		fieldsArr = append(fieldsArr, fieldValue[0].(string))
		placeholdersArr = append(placeholdersArr, fmt.Sprintf("$%d", paramIndex))
		params = append(params, fieldValue[1])
		paramIndex++
	}
	fields = strings.Join(fieldsArr, ", ")
	placeholders = strings.Join(placeholdersArr, ", ")
	return
}
