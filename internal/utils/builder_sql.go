package utils

import (
	"fmt"
	"strings"
)

type Map map[string]interface{}

type BuilderUpdate struct {
	params []interface{}
	paramIndex int
}

func NewBuilderUpdate() *BuilderUpdate {
	return &BuilderUpdate{
        paramIndex: 1,
    }
}

func (b *BuilderUpdate) Build(fieldsValues Map) string{
	fieldsPlaceholders := make([]string, 0, len(fieldsValues))
    for column, value := range fieldsValues {
        fieldsPlaceholders = append(fieldsPlaceholders, fmt.Sprintf("%s = $%d", column, b.paramIndex))
        b.params = append(b.params, value)
        b.paramIndex++
    }
	return strings.Join(fieldsPlaceholders, ", ")
}

func (b *BuilderUpdate) Params() []interface{}{
	return b.params
}

func BuildInsert(fieldsValues Map) (fields string, placeholders string, params []interface{}) {
	fieldsArr := make([]string, 0, len(fieldsValues))
	placeholdersArr := make([]string, 0, len(fieldsValues))
	// params = []interface{}{}
	paramIndex := 1
	for column, value := range fieldsValues {
        fieldsArr = append(fieldsArr, column)
		placeholdersArr = append(placeholdersArr, fmt.Sprintf("$%d", paramIndex))
        params = append(params, value)
        paramIndex++
    }
	fields = strings.Join(fieldsArr, ", ")
	placeholders = strings.Join(placeholdersArr, ", ")
	return
}