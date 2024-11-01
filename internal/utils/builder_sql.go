package utils

import (
	"fmt"
	"strings"
)

type BuilderUpdate struct {
	params []interface{}
	paramIndex int
}
type Map map[string]interface{}

func NewBuilderUpdate() *BuilderUpdate {
	return &BuilderUpdate{
        paramIndex: 1,
    }
}

func (b *BuilderUpdate) Build(fields Map) string{
	setClauses := make([]string, 0, len(fields))
    for column, value := range fields {
        setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, b.paramIndex))
        b.params = append(b.params, value)
        b.paramIndex++
    }
	return strings.Join(setClauses, ", ")
}

func (b *BuilderUpdate) Params() []interface{}{
	return b.params
}