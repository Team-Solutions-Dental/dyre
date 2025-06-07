package endpoint

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/vamuscari/dyre/object"
)

var default_type object.ObjectType = object.STRING_OBJ

type Node interface {
	JSON() string
}

type Service struct {
	Node
	Endpoints map[string]*Endpoint
}

func (s *Service) JSON() string {
	var out bytes.Buffer
	endpointNames := sortedMapKeys(s.Endpoints)
	enpoints := []string{}
	for _, ep := range endpointNames {
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

type Endpoint struct {
	Node
	Service    *Service
	Name       string
	TableName  string
	SchemaName string
	Joins      map[string]Join
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

	fieldNames := sortedMapKeys(e.Fields)
	fields := []string{}

	for _, i := range fieldNames {
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

type Field struct {
	Node
	endpoint  *Endpoint
	Name      string
	FieldType object.ObjectType
	Nullable  bool
}

func (f *Field) Type() object.ObjectType { return f.FieldType }

func (f *Field) Endpoint() *Endpoint { return f.endpoint }
func (f *Field) JSON() string {
	var out bytes.Buffer
	out.WriteString("{")

	out.WriteString(fmt.Sprintf("\"name\" : \"%s\", ", f.Name))
	out.WriteString(fmt.Sprintf("\"type\" : \"%s\", ", f.FieldType))
	out.WriteString(fmt.Sprintf("\"nullable\" : %t ", f.Nullable))

	out.WriteString("}")

	return out.String()
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

// func (j *Join) Endpoint() *Endpoint {
// 	return j.endpoint
// }

func sortedMapKeys[T any](m map[string]T) []string {
	keys := []string{}
	for i := range m {
		keys = append(keys, i)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
