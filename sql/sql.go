package sql

import (
	"fmt"
	"strings"

	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/object/objectRef"
	"github.com/vamuscari/dyre/object/objectType"
)

type Query struct {
	SelectStatements     []SelectStatement
	AliasWhereStatements []string
	TableAlias           string
	Limit                *int
	From                 string
	TableName            string
	WhereStatements      []string
	JoinStatements       []*JoinStatement
	GroupByStatements    []string
	HavingStatements     []string
	OrderBy              []*OrderByStatement
	RefLevel             int
}

func (q *Query) ConstructQuery() string {
	switch q.RefLevel {
	case objectRef.LITERAL:
		return q.tableQuery()
	case objectRef.FIELD:
		return q.tableQuery()
	case objectRef.EXPRESSION:
		return q.aliasQuery()
	case objectRef.GROUP:
		return q.groupQuery()
	default:
		return q.tableQuery()
	}
}

func (q *Query) tableQuery() string {
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

	if len(q.OrderBy) > 0 {
		query = query + orderByConstructor(q.OrderBy)
	}

	return query
}

func (q *Query) aliasQuery() string {
	var query string = "SELECT "

	if q.Limit != nil && *q.Limit > 0 {
		query = query + fmt.Sprintf("TOP %d ", *q.Limit)
	}

	var selectList []string
	for _, v := range q.SelectNameList() {
		selectList = append(selectList, fmt.Sprintf("%s.[%s]", q.TableAlias, v))
	}

	query = query + strings.Join(selectList, ", ")

	query = query + fmt.Sprintf(" FROM ( %s ) AS %s", q.aliasTableQuery(), q.TableAlias)

	if len(q.AliasWhereStatements) > 0 {
		query = query + whereConstructor(q.AliasWhereStatements)
	}

	if len(q.OrderBy) > 0 {
		query = query + orderByConstructor(q.OrderBy)
	}

	return query
}

func (q *Query) aliasTableQuery() string {
	var query string = "SELECT "

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

func (q *Query) groupQuery() string {
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

	if len(q.GroupByStatements) > 0 {
		query = query + groupByConstructor(q.GroupByStatements)
	}

	if len(q.HavingStatements) > 0 {
		query = query + havingConstructor(q.HavingStatements)
	}

	if len(q.OrderBy) > 0 {
		query = query + orderByConstructor(q.OrderBy)
	}

	return query
}

// SELECT * FROM () WHER

func (q *Query) SelectNameList() []string {
	var fields []string
	for _, ss := range q.SelectStatements {
		fields = append(fields, ss.Name())
	}
	return fields
}

// Check for same Name or Alias
func (q *Query) SelectStatementLocation(input string) int {
	names := q.SelectNameList()
	for i, name := range names {
		if name == input {
			return i
		}
	}
	return -1
}

func selectConstructor(stmts []SelectStatement) string {
	var selectStrings []string
	for _, ss := range stmts {
		selectStrings = append(selectStrings, ss.Statement())
	}

	return strings.Join(selectStrings, ", ")
}

// TODO: Type should be enum
type SelectStatement interface {
	Type() string
	ObjectType() objectType.Type
	Name() string
	Statement() string
}

type SelectField struct {
	FieldName *string
	TableName *string
	ObjType   objectType.Type
}

func (sf *SelectField) Type() string                { return "FIELD" }
func (sf *SelectField) ObjectType() objectType.Type { return sf.ObjType }
func (sf *SelectField) Name() string                { return *sf.FieldName }
func (sf *SelectField) Statement() string {
	return fmt.Sprintf("%s.[%s]", *sf.TableName, *sf.FieldName)
}

type SelectExpression struct {
	Expression object.Object
	Alias      *string
}

func (se *SelectExpression) Type() string                { return "EXPRESSION" }
func (se *SelectExpression) ObjectType() objectType.Type { return se.Expression.Type() }
func (se *SelectExpression) Name() string                { return *se.Alias }
func (se *SelectExpression) Statement() string {
	return fmt.Sprintf("( %s ) AS %s", se.Expression.String(), *se.Alias)
}

type SelectGroupField struct {
	FieldName *string
	TableName *string
	ObjType   objectType.Type
}

func (sgf *SelectGroupField) Type() string                { return "GROUP_FIELD" }
func (sgf *SelectGroupField) ObjectType() objectType.Type { return sgf.ObjType }
func (sgf *SelectGroupField) Name() string                { return *sgf.FieldName }
func (sgf *SelectGroupField) Statement() string {
	return fmt.Sprintf("%s.[%s]", *sgf.TableName, *sgf.FieldName)
}

type SelectGroupExpression struct {
	Expression object.Object
	Fn         *string
	Alias      *string
}

func (sge *SelectGroupExpression) Type() string                { return "GROUP_EXPRESSION" }
func (sge *SelectGroupExpression) ObjectType() objectType.Type { return sge.Expression.Type() }
func (sge *SelectGroupExpression) Name() string                { return *sge.Alias }
func (sge *SelectGroupExpression) Statement() string {
	return fmt.Sprintf("%s AS %s", sge.Expression.String(), *sge.Alias)
}

type JoinStatement struct {
	Parent_Query *Query
	Child_Query  *Query
	Parent_On    *string
	Child_On     *string
	JoinType     *string
	Alias        *string
}

func (js *JoinStatement) parentIrOn() string {
	return fmt.Sprintf("%s.[%s]", js.Parent_Query.TableName, *js.Parent_On)
}

func (js *JoinStatement) joinIrOn() string {
	return fmt.Sprintf("%s.[%s]", *js.Alias, *js.Child_On)
}

// TODO: Append select statements from joins
func joinConstructor(joins []*JoinStatement) string {
	var joinArr []string
	for _, j := range joins {

		joinArr = append(joinArr, fmt.Sprintf(" %s JOIN ( %s ) AS %s ON %s = %s", *j.JoinType, j.Child_Query.ConstructQuery(), *j.Alias, j.parentIrOn(), j.joinIrOn()))
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

func havingConstructor(statements []string) string {
	where := ""
	if len(statements) < 1 {
		return where
	}

	if len(statements) == 1 {
		where = fmt.Sprintf(" HAVING %s", statements[0])
		return where
	}
	where = fmt.Sprintf(" HAVING %s", statements[0])
	for i := 1; i < len(statements); i++ {
		where = where + " AND " + statements[i]
	}
	return where
}

func groupByConstructor(statements []string) string {
	var groupByStrings []string
	for _, s := range statements {
		groupByStrings = append(groupByStrings, s)
	}

	return " GROUP BY " + strings.Join(groupByStrings, ", ")
}

type OrderByStatement struct {
	Ascending bool
	FieldName string
}

func orderByConstructor(statements []*OrderByStatement) string {
	if len(statements) < 1 {
		return ""
	}

	var orderByArr []string
	for _, ob := range statements {
		direction := ""
		if ob.Ascending == true {
			direction = " ASC"
		} else {
			direction = " DESC"
		}
		orderByArr = append(orderByArr, ob.FieldName+direction)
	}
	return (" ORDER BY " + strings.Join(orderByArr, ", "))
}
