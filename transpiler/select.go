package transpiler

import (
	"fmt"
)

type selectStatement struct {
	fieldName *string
	tableName *string
	alias     *string
	exclude   bool
}

func (ss *selectStatement) String() string {
	return fmt.Sprintf("%s.%s", *ss.tableName, *ss.fieldName)
}
