package endpoint

import (
	"fmt"
	"github.com/vamuscari/dyre/utils"
	"io"
	"log"
	"os"
	"testing"
)

// TODO: REDO

func TestParseJSON(t *testing.T) {

	const parserJSON string = `
	[
		{
			"name": "ParserTest",
			"tableName": "Table",
			"schemaName": "dbo",
			"fields": [
				{
					"name": "field1",
					"defaultField": true
				},
				{
					"name": "field2",
					"defaultField": false
				},
				"field3"
			]
		}
	]
	`

	dyre_requests, err := ParseJSON([]byte(parserJSON))
	if err != nil {
		t.Fatalf(`error: %v`, err)
	}
	test_name := "ParserTest"
	re, ok := dyre_requests.Endpoints[test_name]
	if !ok {
		t.Error("Map key error for ParserTest")
	}

	t.Run("DyRe_Request.name", func(t *testing.T) {
		want := test_name
		got := re.Name
		if want != got {
			t.Errorf("Want: %s, Got: %s", want, got)
		}
	})

	t.Run("DyRe_Request.FieldNames()", func(t *testing.T) {
		want := []string{"field1", "field2", "field3"}
		got := re.FieldNames
		errors := deepEqualStringArray(want, got)
		if len(errors) > 0 {
			for _, err := range errors {
				t.Error(err)
			}
		}
	})

}

func deepEqualStringArray(arr1 []string, arr2 []string) []error {
	errors := []error{}
	for _, k := range arr1 {
		if !utils.Array_Contains(arr2, k) {
			errors = append(errors, fmt.Errorf("%s not found in %v\n", k, arr2))
		}
	}
	for _, k := range arr2 {
		if !utils.Array_Contains(arr1, k) {
			errors = append(errors, fmt.Errorf("%s not found in %v\n", k, arr1))
		}
	}
	return errors
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
