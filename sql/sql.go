package sql

import (
	"fmt"
	"strings"
)

type Query struct {
	SelectStatements  []*SelectStatement
	Limit             *int
	From              string
	TableName         string
	WhereStatements   []string
	JoinStatements    []*JoinStatement
	GroupByStatements []string
}

func (q *Query) SelectNameList() []string {
	var fields []string
	for _, ss := range q.SelectStatements {
		fields = append(fields, ss.Name())
	}
	return fields
}

type JoinStatement struct {
	Parent_Query *Query
	Child_Query  *Query
	Parent_On    string
	Child_On     string
	JoinType     string
	Alias        string
}

type SelectStatement struct {
	FieldName *string
	TableName *string
	Alias     *string
	Exclude   bool
}

func (ss *SelectStatement) SelectCall() string {
	if ss.Alias != nil {
		return fmt.Sprintf("(%s.[%s]) AS %s", *ss.TableName, *ss.FieldName, *ss.Alias)
	}
	return fmt.Sprintf("%s.[%s]", *ss.TableName, *ss.FieldName)
}

func (ss *SelectStatement) Name() string {
	if ss.Alias != nil {
		return *ss.Alias
	}
	return *ss.FieldName
}

// TODO: Append select statements from joins
func (q *Query) ConstructQuery() string {

	var query string = "SELECT "

	if q.Limit != nil && *q.Limit > 0 {
		query = query + fmt.Sprintf("TOP %d ", *q.Limit)
	}

	query = query + selectConstructor(q.SelectStatements)

	query = query + " FROM " + q.From

	if len(q.JoinStatements) > 0 {
		query = query + joinConstructor(q.JoinStatements)
	}

	if len(q.WhereStatements) > 0 {
		query = query + whereConstructor(q.WhereStatements)
	}

	return query
}

func selectConstructor(selects []*SelectStatement) string {
	var selectStrings []string
	for _, ss := range selects {
		selectStrings = append(selectStrings, ss.SelectCall())
	}

	return strings.Join(selectStrings, ", ")
}

func joinConstructor(joins []*JoinStatement) string {
	var joinArr []string
	for _, j := range joins {

		joinArr = append(joinArr, fmt.Sprintf(" %s JOIN ( %s ) AS %s ON %s = %s", j.JoinType, j.Child_Query.ConstructQuery(), j.Alias, j.parentIrOn(), j.joinIrOn()))
	}

	return strings.Join(joinArr, " ")
}

func whereConstructor(statements []string) string {
	where := ""
	if len(statements) < 1 {
		return where
	}
	if len(statements) == 1 {
		where = fmt.Sprintf(" WHERE %s", statements[0])
		return where
	}
	where = fmt.Sprintf(" WHERE %s", statements[0])
	for i := 1; i < len(statements); i++ {
		where = where + " AND " + statements[i]
	}
	return where
}

func (q *Query) SelectStatementLocation(input string) int {
	for i, ss := range q.SelectStatements {
		if *ss.FieldName == input {
			return i
		}
	}
	return -1
}

func (js *JoinStatement) parentIrOn() string {
	return fmt.Sprintf("%s.[%s]", js.Parent_Query.TableName, js.Parent_On)
}

func (js *JoinStatement) joinIrOn() string {
	return fmt.Sprintf("%s.[%s]", js.Child_Query.TableName, js.Child_On)
}
