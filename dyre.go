package dyre

import (
	"database/sql"
	"errors"
	"fmt"
	"maps"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/parser"
)

// TODO: sql tools

type DyRe_Field struct {
	name      string
	typeName  string
	sqlType   string
	defaultField bool
	sqlSelect string
	tag       map[string]string
}

type DyRe_Constructor struct {
	_request   *DyRe_Request
	_ast       *ast.Query
	_headers   []string
	_sqlFields []string
	_sqlTypes  []string
	_fields    []DyRe_Field
	_tableName string
	_limit     int
}

type DyRe_Request struct {
	name        string
	requestType string
	fields      map[string]DyRe_Field
	fieldNames  []string
	tableName   string
}

var DefaultType = "sql.NullString"
var Types = map[string]interface{}{
	"string":          string(""),
	"int":             int(0),
	"int8":            int8(0),
	"int16":           int16(0),
	"int32":           int32(0),
	"int64":           int64(0),
	"uint":            uint(0),
	"uint8":           uint8(0), // byte
	"uint16":          uint16(0),
	"uint32":          uint32(0),
	"uint64":          uint64(0),
	"uintptr":         uintptr(0), // rune
	"bool":            bool(false),
	"float32":         float32(0),
	"float64":         float64(0),
	"complex64":       complex64(0),
	"complex128":      complex128(0),
	"sql.NullString":  sql.NullString{},
	"sql.NullBool":    sql.NullBool{},
	"sql.NullByte":    sql.NullByte{},
	"sql.NullTime":    sql.NullTime{},
	"sql.NullInt16":   sql.NullInt16{},
	"sql.NullInt32":   sql.NullInt32{},
	"sql.NullInt64":   sql.NullInt64{},
	"sql.NullFloat64": sql.NullFloat64{},
}

// TODO:
// Parse Query parameters into ast and assign to corresponding fields for building request.
func (re *DyRe_Request) ParseQuery(query string) (DyRe_Constructor, error) {
	var constructor DyRe_Constructor

	tokens := lexer.New(query)
	q := parser.New(tokens)
	ast := q.ParseQuery()
	if len(q.Errors()) > 0 {
		return constructor, errors.New("Parser Error")
	}

	constructor._ast = ast

	// TODO: return default fields
	// construct statement?
	if len(ast.Statements) == 0 {

	}

	// eval
	for _, stmt := range ast.Statements {
		if stmt.TokenLiteral
		

	}

	return constructor, nil
}

// Validates incoming fields and groups.
// Returns a validated struct for making sql queries.
// if group is found dont check fields for group field match to avoid duplication
func (re *DyRe_Request) ValidateRequest(fields []string) (DyRe_Constructor, error) {
	var selected DyRe_Constructor
	for _, re_field := range re.fields {
		if re_field.required == true || contains(fields, re_field.name) {
			selected._fields = append(selected._fields, re_field)
			selected._sqlFields = append(selected._sqlFields, re_field.sqlSelect)
			selected._headers = append(selected._headers, re_field.name)
			selected._sqlTypes = append(selected._sqlTypes, re_field.typeName)
		}
	}

	if len(selected._fields) == 0 {
		return selected, errors.New(fmt.Sprintf("No valid fields or groups selected for %s", re.name))
	}

	selected._request = re

	return selected, nil
}

func (re *DyRe_Request) FieldNames() []string {
	return re.fieldNames
}

func (re *DyRe_Request) TableName() string {
	return re.tableName
}

func (re *DyRe_Request) Fields() map[string]DyRe_Field {
	return maps.Clone(re.fields)
}

// Returns the list of names for the fields that were calles
func (con *DyRe_Constructor) Headers() []string {
	headers := []string{}
	for _, v := range con._headers {
		headers = append(headers, v)
	}
	return headers
}

// Fields for SQL Select.
//
// SELECT  {Fields} FROM {Table}
func (con *DyRe_Constructor) SQLFields() []string {
	sqlFields := []string{}
	for _, v := range con._sqlFields {
		sqlFields = append(sqlFields, v)
	}
	return sqlFields
}

func (con *DyRe_Constructor) ConstructStatement() string {
	SELECT
	FROM
	WHERE

}

func (field *DyRe_Field) Name() string {
	return field.name
}

func (field *DyRe_Field) Required() bool {
	return field.required
}

func (field *DyRe_Field) SQLSelect() string {
	return field.sqlSelect
}

func (field *DyRe_Field) Type() string {
	return field.typeName
}

func contains(a []string, l string) bool {
	for _, v := range a {
		if l == v {
			return true
		}
	}
	return false
}
