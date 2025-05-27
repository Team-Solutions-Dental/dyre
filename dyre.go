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

func (d *Dyre) Request(req string, query string) (*transpiler.IR, error) {
	endpoint, ok := d.service.Endpoints[req]
	if !ok {
		return nil, errors.New("Invalid Endpoint. got=" + req)
	}

	return transpiler.New(query, endpoint), nil
}
