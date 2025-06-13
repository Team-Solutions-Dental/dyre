package dyre

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/transpiler"
)

func Init(filepath string) Dyre {
	var dyre Dyre

	json_bytes := openDyreJSON(filepath)
	service, err := endpoint.ParseJSON(json_bytes)
	if err != nil {
		log.Panic(err)
	}

	dyre.service = service

	return dyre
}

func openDyreJSON(path string) []byte {
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Panic(err)
		return nil
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Panic(err)
		return nil
	}

	return byteValue

}

type Dyre struct {
	service *endpoint.Service
}

func (d *Dyre) Endpoint(req string) (*Endpoint, error) {
	endpoint, ok := d.service.Endpoints[req]
	if !ok {
		return nil, errors.New("Invalid Endpoint. got=" + req)
	}

	return &Endpoint{ref: endpoint}, nil
}

func (d *Dyre) Request(req string, query string) (*transpiler.PrimaryIR, error) {
	endpoint, ok := d.service.Endpoints[req]
	if !ok {
		return nil, errors.New("Invalid Endpoint. got=" + req)
	}

	return transpiler.New(query, endpoint)
}

func (d *Dyre) EndpointNames() []string {
	var names []string
	for k := range d.service.Endpoints {
		names = append(names, k)
	}

	return names
}

func (d *Dyre) AllEndpointPaths(depth int) [][]string {
	return d.service.AllEndpointPaths(depth)
}

type Endpoint struct {
	ref *endpoint.Endpoint
}

func (e *Endpoint) Request(query string) (*transpiler.PrimaryIR, error) {
	return transpiler.New(query, e.ref)
}

func (e *Endpoint) Fields() []string {
	return e.ref.FieldNames
}

func (e *Endpoint) Joins() []string {
	var joins []string
	for _, j := range e.ref.Joins {
		joins = append(joins, j.Name())
	}
	return joins
}
