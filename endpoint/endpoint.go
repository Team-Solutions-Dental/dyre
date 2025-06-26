package endpoint

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/vamuscari/dyre/object/objectType"
	"github.com/vamuscari/dyre/sql"
	"github.com/vamuscari/dyre/utils"
)

var default_type objectType.Type = objectType.STRING

type Node interface {
	JSON() string
}

type Service struct {
	Node
	Endpoints     map[string]*Endpoint
	EndpointNames []string
	Settings      Settings
}

func (s *Service) JSON() string {
	var out bytes.Buffer
	enpoints := []string{}
	for _, ep := range s.EndpointNames {
		enpoints = append(enpoints, s.Endpoints[ep].JSON())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(enpoints, ", "))
	out.WriteString("]")

	return out.String()
}

func (s *Service) GetEndpoint(req string) (*Endpoint, error) {
	if s.Endpoints == nil {
		return nil, errors.New("No Endpoints Available " + req)
	}

	endpoint, ok := s.Endpoints[req]
	if !ok {
		return nil, errors.New("Invalid Endpoint. got=" + req)
	}
	return endpoint, nil
}

// ["Customers"]
// ["Customers", "Invoices"]
func (s *Service) AllEndpointPaths(depth int) [][]string {
	var endpoints [][]string
	for _, i := range s.EndpointNames {
		endpoint := s.Endpoints[i]
		new_path := []string{}
		endpoints = append(endpoints, EndpointPaths(endpoint, new_path, 0, depth)...)
	}

	return endpoints
}

func EndpointPaths(endpoint *Endpoint, currpath []string, depth int, depthStop int) [][]string {
	path := append(currpath, endpoint.Name)
	subPaths := [][]string{}
	subPaths = append(subPaths, []string{endpoint.Name})

	if depth >= depthStop {
		return subPaths
	}

	if len(endpoint.Joins) < 1 {
		return subPaths
	}

	for _, j := range endpoint.JoinNames {
		if utils.Array_Contains(path, j) {
			continue
		}

		join := endpoint.Joins[j]

		for _, p := range EndpointPaths(join.ChildEndpoint(), path, (depth + 1), depthStop) {
			new_path := append([]string{endpoint.Name}, p...)
			subPaths = append(subPaths, new_path)
		}
	}

	return subPaths
}

type Endpoint struct {
	Node
	Service    *Service
	Name       string
	TableName  string
	SchemaName string
	Joins      map[string]Join
	JoinNames  []string
	Fields     map[string]Field
	FieldNames []string
}

func (e *Endpoint) JSON() string {
	var out bytes.Buffer
	out.WriteString("{ ")

	joins := []string{}
	for _, join := range e.Joins {
		joins = append(joins, join.JSON())
	}

	fields := []string{}
	for _, i := range e.FieldNames {
		field := e.Fields[i]
		fields = append(fields, field.JSON())
	}

	out.WriteString(fmt.Sprintf("\"name\" : \"%s\", ", e.Name))
	out.WriteString(fmt.Sprintf("\"tableName\" : \"%s\", ", e.TableName))
	out.WriteString(fmt.Sprintf("\"schemaName\" : \"%s\", ", e.SchemaName))
	out.WriteString("\"joins\" : [")
	out.WriteString(strings.Join(joins, ", "))
	out.WriteString("],")
	out.WriteString("\"fields\" : [")
	out.WriteString(strings.Join(fields, ", "))
	out.WriteString("]")

	out.WriteString("}")

	return out.String()
}

func (e *Endpoint) TS() string {
	var out bytes.Buffer
	out.WriteString("interface " + e.Name + " { ")
	for _, f := range e.FieldNames {
		field := e.Fields[f]
		out.WriteString("\n  ")
		out.WriteString(field.TS())
	}
	out.WriteString("\n}")

	return out.String()
}

type Field struct {
	Node
	Endpoint  *Endpoint
	Name      string
	FieldType objectType.Type
	Nullable  bool
}

func (f *Field) Type() objectType.Type { return f.FieldType }

func (f *Field) JSON() string {
	var out bytes.Buffer
	out.WriteString("{")

	out.WriteString(fmt.Sprintf("\"name\" : \"%s\", ", f.Name))
	out.WriteString(fmt.Sprintf("\"type\" : \"%s\", ", f.FieldType))
	out.WriteString(fmt.Sprintf("\"nullable\" : %t ", f.Nullable))

	out.WriteString("}")

	return out.String()
}

func (f *Field) TS() string {
	var out bytes.Buffer
	out.WriteString(f.Name)
	if f.Nullable {
		out.WriteString("?")
	}
	out.WriteString(": ")

	out.WriteString(objTypeToTsType(f.FieldType))
	out.WriteString(";")

	return out.String()
}

func (f *Field) SelectStatement() *sql.SelectField {
	return &sql.SelectField{
		FieldName: &f.Name,
		TableName: &f.Endpoint.TableName,
	}
}

// Maybe reference pointer to field?
type Join struct {
	Node
	parentEndpoint    *Endpoint
	childEndpoint     *Endpoint
	childEndpointName string
	Parent_ON         string
	Child_ON          string
}

func (j *Join) JSON() string {
	var out bytes.Buffer
	out.WriteString("{")

	out.WriteString(
		fmt.Sprintf("\"endpoint\" : \"%s\", ", j.childEndpointName),
	)
	out.WriteString(
		fmt.Sprintf("\"on\": [\"%s\",\"%s\"] ", j.Parent_ON, j.Child_ON),
	)

	out.WriteString("}")

	return out.String()
}

func (j *Join) Name() string {
	return j.childEndpointName
}

func (j *Join) ParentEndpoint() *Endpoint {
	return j.parentEndpoint
}

func (j *Join) ChildEndpoint() *Endpoint {
	return j.childEndpoint
}

func objTypeToTsType(obj objectType.Type) string {
	switch obj {
	case objectType.STRING:
		return "string"
	case objectType.BOOLEAN:
		return "boolean"
	case objectType.INTEGER:
		return "number"
	case objectType.FLOAT:
		return "number"
	case objectType.DATETIME:
		return "Date"
	case objectType.DATE:
		return "Date"
	case objectType.NULL:
		return "null"
	default:
		return "string"
	}
}

type Settings struct {
	BracketedColumns bool
}
