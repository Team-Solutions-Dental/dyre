package endpoint

import (
	"errors"
	"strings"
)

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

func (ep *Endpoint) DefaultRequest() string {
	var request []string
	for _, f := range ep.Fields {
		if f.DefaultField {
			request = append(request, f.Name+":,")
		}
	}
	return strings.Join(request, "")
}

type Field struct {
	endpoint     *Endpoint
	Name         string
	DefaultField bool
}

func (f *Field) Endpoint() *Endpoint {
	return f.endpoint
}
