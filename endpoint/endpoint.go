package endpoint

import (
	"errors"

	"github.com/vamuscari/dyre/object"
)

var default_type object.ObjectType = object.STRING_OBJ

type Service struct {
	Endpoints map[string]*Endpoint
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
	Service    *Service
	Name       string
	Fields     map[string]Field
	FieldNames []string
	TableName  string
	SchemaName string
}

type Field struct {
	endpoint  *Endpoint
	Name      string
	FieldType object.ObjectType
}

func (f *Field) Type() object.ObjectType { return f.FieldType }

func (f *Field) Endpoint() *Endpoint {
	return f.endpoint
}
